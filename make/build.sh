#!/bin/bash

#make.sh {build} {shell|atlas|titan} {staging|prod} {descr} {name} {manu}"
#make.sh build  appl   envr    desc   name    manu"
#        1      2      3       4      5       6
#make.sh build  shell staging 'desc' teva    teva"
#make.sh build  shell prod    'desc' teva    teva"   # is a manu
#make.sh build  shell prod    'desc' modeln  bayer"  # is a proc
#make.sh build  shell staging 'desc' amgen   amgen"
#make.sh build  atlas staging 'cntr' brg     bayer"  # is a proc (but brg can access all)
#make.sh build  atlas staging 'cntr' brg     brg"    # is a manu (but brg can access all)

#make.sh {deploy} {shell|atlas|service} {staging|prod} {descr} {name} {readme}"
#make.sh deploy appl   envr    desc      name    readme"
#        1      2      3       4         5       6
#make.sh deploy shell prod    'version' amgen   readme.md"
#make.sh deploy atlas staging 'desc'    myImg"

if [ "$#" -ne 6 ]; then
    echo "Illegal number of parameters ($#)"
    echo "build.sh {build|deploy} {shell|atlas|titan} {staging|prod} {descr} {name} {manu|proc|container_name}"
    exit
fi

cd ..

MYOS=macos
MYARCH=arm

prdt="binary-v5"
nspc="v5"

comd="$1"   # Build or deploy
appl="$2"   # Shell, atlas or titan
envr="$3"   # Dev, devint, staging, prod, etc.
desc="$4"   # Description - embedded in all apps, and in shell/deploy (binary) (for servers it's the container id/version/name)
name="$5"   # Identity, like brg or amgen
manu="$6"   # Manufacturer. If the same as name/$5, then the type is manu, else type is proc.

if [ "$appl" == "shell" ]; then
  if [ "$manu" == "$name" ]; then
    type="manu"
  else
    type="proc"
    manu=""
  fi
elif [ "$appl" == "atlas" ]; then
  cntr=$desc
elif [ "$appl" == "titan" ]; then
  name="brg"
  cntr=$desc
fi

make_cert() {
local appl=$1
local addr=$2
export X509_O=${name}.secondsightsolutions.com  # brg.secondsightsolutions.com
export X509_OU=${appl}.${envr}.${prdt}          # atlas.staging.binary-v5
export X509_CN=${name}                          # brg
export X509_EM=${prdt}@${X509_O}
rm -f cert-ext.txt ca-cert.srl *.b64 *.pem
az keyvault secret download --vault-name v5-atlas-vault -n ca-cert -f ca-cert.pem.b64
az keyvault secret download --vault-name v5-atlas-vault -n ca-pkey -f ca-pkey.pem.b64
base64 -d -i ca-cert.pem.b64 > ca-cert.pem
base64 -d -i ca-pkey.pem.b64 > ca-pkey.pem
echo "subjectAltName=IP:0.0.0.0,IP:127.0.0.1,IP:${addr}" > cert-ext.txt
extcmd="-extfile cert-ext.txt"
openssl req -newkey rsa:4096 -nodes -keyout pkey.pem -out req.pem -subj "/C=US/ST=DC/L=DC/O=$X509_O/OU=$X509_OU/CN=$X509_CN/emailAddress=$X509_EM" > /dev/null 2>&1
openssl x509 -req -in req.pem -days 3650 -sha256 -CA ca-cert.pem -CAkey ca-pkey.pem -CAcreateserial -out cert.pem $extcmd > /dev/null 2>&1
openssl verify -CAfile ca-cert.pem cert.pem
./make/crypt --phrase=${V5_APPL_PHRS} --encrypt=pkey.pem --output=pkey.pem.b64
base64 -i cert.pem > cert.pem.b64
}
export V5_APPL_CACR="$(az keyvault secret show --vault-name v5-atlas-vault  --name ca-cert | jq .value | tr -d '"')"
export V5_APPL_PKEY="$(az keyvault secret show --vault-name v5-atlas-vault  --name ca-pkey | jq .value | tr -d '"')"
export V5_APPL_PHRS="$(az keyvault secret show --vault-name v5-atlas-vault  --name phrase  | jq .value | tr -d '"')"
export V5_APPL_SALT="$(az keyvault secret show --vault-name v5-atlas-vault  --name salt    | jq .value | tr -d '"')"
export V5_APPL_ENVR=$envr
export V5_APPL_VERS="$(date '+%s')"
export V5_APPL_HASH=hash=$(git rev-parse --short HEAD)

if [ "$name" == "brg" ] || [ "$appl" == "atlas" ];then
export V5_ATLAS_HOST="$(az appconfig kv show -n v5-atlas-config --key atlas_db_host --label ${envr} | jq .value | tr -d '"')"
export V5_ATLAS_PORT="$(az appconfig kv show -n v5-atlas-config --key atlas_db_port --label ${envr} | jq .value | tr -d '"')"
export V5_ATLAS_NAME="$(az appconfig kv show -n v5-atlas-config --key atlas_db_name --label ${envr} | jq .value | tr -d '"')"
export V5_ATLAS_USER="$(az appconfig kv show -n v5-atlas-config --key atlas_db_user --label ${envr} | jq .value | tr -d '"')"
export V5_ATLAS_PASS="$(az keyvault secret show --vault-name v5-atlas-vault --name aadmin | jq .value | tr -d '"')"
export V5_ATLAS_GRPC="$(az appconfig kv show -n v5-atlas-config --key atlas_grpc    --label ${envr} | jq .value | tr -d '"')"
export V5_TITAN_GRPC="$(az appconfig kv show -n v5-atlas-config --key titan_grpc    --label ${envr} | jq .value | tr -d '"')"
export V5_ATLAS_OGTM="$(az appconfig kv show -n v5-atlas-config --key atlas_ogtm    --label ${envr} | jq .value | tr -d '"')"
export V5_ATLAS_OGKY="$(az appconfig kv show -n v5-atlas-config --key atlas_ogky    --label ${envr} | jq .value | tr -d '"')"
make_cert atlas ${V5_ATLAS_GRPC}
export V5_ATLAS_CERT="$(cat cert.pem.b64)"
export V5_ATLAS_PKEY="$(cat pkey.pem.b64)"
fi

if [ "$name" == "brg" ] || [ "$appl" == "titan" ];then
export V5_TITAN_HOST="$(az appconfig kv show -n v5-atlas-config --key titan_db_host --label ${envr} | jq .value | tr -d '"')"
export V5_TITAN_PORT="$(az appconfig kv show -n v5-atlas-config --key titan_db_port --label ${envr} | jq .value | tr -d '"')"
export V5_TITAN_NAME="$(az appconfig kv show -n v5-atlas-config --key titan_db_name --label ${envr} | jq .value | tr -d '"')"
export V5_TITAN_USER="$(az appconfig kv show -n v5-atlas-config --key titan_db_user --label ${envr} | jq .value | tr -d '"')"
export V5_TITAN_PASS="$(az keyvault secret show --vault-name v5-atlas-vault --name aadmin | jq .value | tr -d '"')"
export V5_CITUS_HOST="$(az appconfig kv show -n v5-atlas-config --key citus_db_host --label ${envr} | jq .value | tr -d '"')"
export V5_CITUS_PORT="$(az appconfig kv show -n v5-atlas-config --key citus_db_port --label ${envr} | jq .value | tr -d '"')"
export V5_CITUS_NAME="$(az appconfig kv show -n v5-atlas-config --key citus_db_name --label ${envr} | jq .value | tr -d '"')"
export V5_CITUS_USER="$(az appconfig kv show -n v5-atlas-config --key citus_db_user --label ${envr} | jq .value | tr -d '"')"
export V5_CITUS_PASS="$(az keyvault secret show --vault-name v5-atlas-vault --name citus | jq .value | tr -d '"')"
export V5_ESPDB_HOST="$(az appconfig kv show -n v5-atlas-config --key esp_db_host   --label ${envr} | jq .value | tr -d '"')"
export V5_ESPDB_PORT="$(az appconfig kv show -n v5-atlas-config --key esp_db_port   --label ${envr} | jq .value | tr -d '"')"
export V5_ESPDB_NAME="$(az appconfig kv show -n v5-atlas-config --key esp_db_name   --label ${envr} | jq .value | tr -d '"')"
export V5_ESPDB_USER="$(az appconfig kv show -n v5-atlas-config --key esp_db_user   --label ${envr} | jq .value | tr -d '"')"
export V5_ESPDB_PASS="$(az keyvault secret show --vault-name v5-atlas-vault --name espdb | jq .value | tr -d '"')"
export V5_TITAN_GRPC="$(az appconfig kv show -n v5-atlas-config --key titan_grpc    --label ${envr} | jq .value | tr -d '"')"
export V5_TITAN_OGTM="$(az appconfig kv show -n v5-atlas-config --key titan_ogtm    --label ${envr} | jq .value | tr -d '"')"
export V5_TITAN_OGKY="$(az appconfig kv show -n v5-atlas-config --key titan_ogky    --label ${envr} | jq .value | tr -d '"')"
make_cert titan ${V5_TITAN_GRPC}
export V5_TITAN_CERT="$(cat cert.pem.b64)"
export V5_TITAN_PKEY="$(cat pkey.pem.b64)"
fi

if [ "$name" == "brg" ] || [ "$appl" == "shell" ];then
export V5_ATLAS_GRPC="$(az appconfig kv show -n v5-atlas-config --key atlas_grpc --label ${envr} | jq .value | tr -d '"')"
make_cert shell 127.0.0.1
export V5_SHELL_CERT="$(cat cert.pem.b64)"
export V5_SHELL_PKEY="$(cat pkey.pem.b64)"
fi

echo -n $name          > embed/name.txt
echo -n $appl          > embed/appl.txt
echo -n $type          > embed/type.txt
echo -n $manu          > embed/manu.txt
echo -n $V5_APPL_ENVR  > embed/envr.txt
echo -n $V5_APPL_CACR  > embed/cacr.txt
echo -n $V5_APPL_SALT  > embed/salt.txt
echo -n $V5_APPL_PHRS  > embed/phrs.txt
echo -n $V5_APPL_VERS  > embed/vers.txt
echo -n $V5_APPL_HASH  > embed/hash.txt
echo -n $V5_APPL_DESC  > embed/desc.txt

if [ "$name" == "brg" ] || [ "$appl" == "atlas" ];then
echo -n $V5_ATLAS_CERT > embed/atlas_cert.txt
echo -n $V5_ATLAS_PKEY > embed/atlas_pkey.txt
echo -n $V5_ATLAS_OGKY > embed/atlas_ogky.txt
echo -n $V5_ATLAS_OGTM > embed/atlas_ogtm.txt
echo -n $V5_ATLAS_GRPC > embed/atlas_grpc.txt
echo -n $V5_TITAN_GRPC > embed/titan_grpc.txt
echo -n $V5_ATLAS_HOST > embed/atlas_host.txt
echo -n $V5_ATLAS_PORT > embed/atlas_port.txt
echo -n $V5_ATLAS_NAME > embed/atlas_name.txt
echo -n $V5_ATLAS_USER > embed/atlas_user.txt
echo -n $V5_ATLAS_PASS > embed/atlas_pass.txt
fi

if [ "$name" == "brg" ] || [ "$appl" == "titan" ];then
echo -n $V5_TITAN_CERT > embed/titan_cert.txt
echo -n $V5_TITAN_PKEY > embed/titan_pkey.txt
echo -n $V5_TITAN_OGKY > embed/titan_ogky.txt
echo -n $V5_TITAN_OGTM > embed/titan_ogtm.txt
echo -n $V5_TITAN_GRPC > embed/titan_grpc.txt
echo -n $V5_CITUS_HOST > embed/citus_host.txt
echo -n $V5_CITUS_PORT > embed/citus_port.txt
echo -n $V5_CITUS_NAME > embed/citus_name.txt
echo -n $V5_CITUS_USER > embed/citus_user.txt
echo -n $V5_CITUS_PASS > embed/citus_pass.txt
echo -n $V5_ESPDB_HOST > embed/espdb_host.txt
echo -n $V5_ESPDB_PORT > embed/espdb_port.txt
echo -n $V5_ESPDB_NAME > embed/espdb_name.txt
echo -n $V5_ESPDB_USER > embed/espdb_user.txt
echo -n $V5_ESPDB_PASS > embed/espdb_pass.txt
echo -n $V5_TITAN_HOST > embed/titan_host.txt
echo -n $V5_TITAN_PORT > embed/titan_port.txt
echo -n $V5_TITAN_NAME > embed/titan_name.txt
echo -n $V5_TITAN_USER > embed/titan_user.txt
echo -n $V5_TITAN_PASS > embed/titan_pass.txt
fi

if [ "$name" == "brg" ] || [ "$appl" == "shell" ];then
echo -n $V5_SHELL_CERT > embed/shell_cert.txt
echo -n $V5_SHELL_PKEY > embed/shell_pkey.txt
echo -n $V5_ATLAS_GRPC > embed/atlas_grpc.txt
fi

rm -rf run
mkdir run
CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -o run/${prdt}_amd_linux
CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build -o run/${prdt}_amd_macos
CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build -o run/${prdt}_arm_macos
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o run/${prdt}_amd_windows

echo -n "" > embed/name.txt
echo -n "" > embed/appl.txt
echo -n "" > embed/cacr.txt
echo -n "" > embed/pkey.txt
echo -n "" > embed/salt.txt
echo -n "" > embed/phrs.txt
echo -n "" > embed/type.txt
echo -n "" > embed/vers.txt
echo -n "" > embed/hash.txt
echo -n "" > embed/manu.txt
echo -n "" > embed/desc.txt

echo -n "" > embed/shell_cert.txt
echo -n "" > embed/shell_pkey.txt
echo -n "" > embed/atlas_cert.txt
echo -n "" > embed/atlas_pkey.txt
echo -n "" > embed/titan_cert.txt
echo -n "" > embed/titan_pkey.txt

echo -n "" > embed/atlas_grpc.txt
echo -n "" > embed/titan_grpc.txt

echo -n "" > embed/atlas_grpc.txt
echo -n "" > embed/atlas_ogky.txt
echo -n "" > embed/atlas_ogtm.txt
echo -n "" > embed/titan_ogky.txt
echo -n "" > embed/titan_ogtm.txt
echo -n "" > embed/titan_grpc.txt

echo -n "" > embed/atlas_host.txt
echo -n "" > embed/atlas_port.txt
echo -n "" > embed/atlas_name.txt
echo -n "" > embed/atlas_user.txt
echo -n "" > embed/atlas_pass.txt
echo -n "" > embed/citus_host.txt
echo -n "" > embed/citus_port.txt
echo -n "" > embed/citus_name.txt
echo -n "" > embed/citus_user.txt
echo -n "" > embed/citus_pass.txt
echo -n "" > embed/espdb_host.txt
echo -n "" > embed/espdb_port.txt
echo -n "" > embed/espdb_name.txt
echo -n "" > embed/espdb_user.txt
echo -n "" > embed/espdb_pass.txt
echo -n "" > embed/titan_host.txt
echo -n "" > embed/titan_port.txt
echo -n "" > embed/titan_name.txt
echo -n "" > embed/titan_user.txt
echo -n "" > embed/titan_pass.txt

envsubst < make/set.env > run.sh
echo "" >> run.sh
echo "./run/${prdt}_${MYARCH}_${MYOS} \$@" >> run.sh
chmod ugo+x run.sh

#kubectl config use-context sss-svcs-${envr}-k8s

#if [ "$appl" == "titan" ]; then
#  export PORT=$port
#  export APP=$nspc
#  cp ../make/service.yaml . > /dev/null 2>&1
#  envsubst < service.yaml | kubectl apply -f -
#  export svip=$(kubectl -n ${nspc} get services ${nspc}-service -o=jsonpath='{.status.loadBalancer.ingress[0].ip}')
#  echo "Setting service IP $svip in 1Pass"
#  op item edit "deployments" $envr.BIN_V5_HOST=$svip --vault=Configuration > /dev/null
#  serv="$(op read op://Configuration/deployments/${envr}/BIN_V5_HOST)"
#fi

# rm -f cert-ext.txt cert.pem pkey.pem pkey.pem.b64 req.pem ca-*.*
# rm -f deployment.yaml service.yaml Dockerfile Dockerfile.run

rm -f cert-ext.txt ca-cert.srl *.b64 *.pem
cd make
