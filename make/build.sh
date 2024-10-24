#!/bin/bash

#make.sh {build} {client|server|service} {staging|prod} {descr} {name} {manu}"
#make.sh build  whch   envr    desc   name    manu"
#        1      2      3       4      5       6
#make.sh build  client staging 'desc' teva    teva"
#make.sh build  client prod    'desc' teva    teva"   # is a manu
#make.sh build  client prod    'desc' modeln  bayer"  # is a proc
#make.sh build  client staging 'desc' amgen   amgen"
#make.sh build  server staging 'cntr' brg     bayer"  # is a proc (but brg can access all)
#make.sh build  server staging 'cntr' brg     brg"    # is a manu (but brg can access all)

#make.sh {deploy} {client|server|service} {staging|prod} {descr} {name} {readme}"
#make.sh deploy whch   envr    desc      name    readme"
#        1      2      3       4         5       6
#make.sh deploy client prod    'version' amgen   readme.md"
#make.sh deploy server staging 'desc'    myImg"

if [ "$#" -ne 6 ]; then
    echo "Illegal number of parameters ($#)"
    echo "build.sh {build|deploy} {client|server|service} {staging|prod} {descr} {name} {manu|proc|container_name}"
    exit
fi

cd ..

MYOS=macos
MYARCH=arm

appl="binary-v5"
nspc="bin"
base="340B_ESP_binary"
srvh="127.0.0.1"  # server host (server-side binary)
srvp="23460"      # server port
svch="127.0.0.1"  # service host (formerly the rebate binary server, hosted in SSS)
svcp="23461"      # service port
host=""           # host embedded in X509 certificate
port=""           # port embedded in X509 certificate

comd="$1"   # Build or deploy
whch="$2"   # Server, client or service
envr="$3"   # Dev, devint, staging, prod, etc.
desc="$4"   # Description - embedded in all apps, and in client/deploy (binary) (for servers it's the container id/version/name)
name="$5"   # Identity, like brg or amgen
manu="$6"   # Manufacturer. If the same as name/$5, then the type is manu, else type is proc.

vers="$(date '+%s')"
hash=$(git rev-parse --short HEAD)

all="op://Configuration/all/all"
cfg="op://Configuration/deployments/${envr}"

type=""
cntr=""
prfx=${appl}_${whch}
phrs="$(op read ${all}/SSS_PASSPHRASE)"
salt="$(op read ${all}/SSS_SALT_ENCR_B64)"
team="$(op read ${cfg}/OG_TEAM)"
apik="$(op read ${cfg}/OG_APIK)"
ogpr="$(op read ${cfg}/OG_PING_ROUTE)"
azbl="$(op read ${cfg}/BIN_BLOB_ACCT)"
azky="$(op read ${cfg}/BIN_BLOB_KEY)"
gray="$(op read ${cfg}/GRAYLOG)"
mypk=""
mycr=""
cacr=""

if [ "$whch" == "client" ]; then
  if [ "$manu" == "$name" ]; then
    type="manu"
  else
    type="proc"
    manu=""
  fi
elif [ "$whch" == "server" ]; then
  host="$srvh"
  port="$srvp"
  cntr=$desc
elif [ "$whch" == "service" ]; then
  host="$svch"
  port="$svcp"
  name="brg"
  cntr=$desc
fi

echo "comd=$comd"
echo "whch=$whch"
echo "envr=$envr"
echo "desc=$desc"
echo "name=$name"
echo "appl=$appl"
echo "nspc=$nspc"
echo "srvh=$srvh"
echo "srvp=$srvp"
echo "svch=$svch"
echo "svcp=$svcp"
echo "base=$base"
echo "manu=$manu"
echo "vers=$vers"
echo "hash=$hash"
echo "host=$host"
echo "port=$port"
echo "type=$type"
echo "cntr=$cntr"
echo "prfx=$prfx"
echo "phrs=$phrs"
echo "salt=$salt"
echo "team=$team"
echo "apik=$apik"
echo "ogpr=$ogpr"
echo "azbl=$azbl"
echo "azky=$azky"
echo "gray=$gray"

rm -f cert-ext.txt ca-*.* cert.pem pkey.pem pkey.pem.b64 req.pem
echo "$(op read ${all}/SSS_CA_CERT_PEM)" > ca-cert.pem
echo "$(op read ${all}/SSS_CA_PKEY_PEM)" > ca-pkey.pem
export X509_O=${name}.secondsightsolutions.com  # brg.secondsightsolutions.com
export X509_OU=${whch}.${envr}.${appl}          # client.staging.binary-v5
export X509_CN=${name}                          # brg
export X509_EM=${appl}@${X509_O}
echo "X_O =$X509_O"
echo "X_OU=$X509_OU"
echo "X_CN=$X509_CN"
echo "X_EM=$X509_EM"
if [ "$whch" == "server" ]; then
  echo "subjectAltName=IP:0.0.0.0,IP:127.0.0.1,IP:${host}" > cert-ext.txt
  extcmd="-extfile cert-ext.txt"
fi
openssl req -newkey rsa:4096 -nodes -keyout pkey.pem -out req.pem -subj "/C=US/ST=DC/L=DC/O=$X509_O/OU=$X509_CN/CN=$X509_CN/emailAddress=$X509_EM" > /dev/null 2>&1
openssl x509 -req -in req.pem -days 3650 -sha256 -CA ca-cert.pem -CAkey ca-pkey.pem -CAcreateserial -out cert.pem $extcmd > /dev/null 2>&1
openssl verify -CAfile ca-cert.pem cert.pem
./make/crypt --phrase=${phrs} --encrypt=pkey.pem --output=pkey.pem.b64
mypk="$(cat pkey.pem.b64)"
mycr="$(base64 -i cert.pem)"
cacr="$(base64 -i ca-cert.pem)"

echo -n $whch > embed/appl.txt
echo -n $srvh > embed/srvh.txt
echo -n $svch > embed/svch.txt
echo -n $cacr > embed/cacr.txt
echo -n $mycr > embed/mycr.txt
echo -n $mypk > embed/pkey.txt
echo -n $salt > embed/salt.txt
echo -n $phrs > embed/phrs.txt
echo -n $type > embed/type.txt
echo -n $name > embed/name.txt
echo -n $vers > embed/vers.txt
echo -n $hash > embed/hash.txt
echo -n $desc > embed/desc.txt
echo -n $envr > embed/envr.txt
if [ "$type" == "manu" ]; then
  echo -n $manu > embed/manu.txt
fi

mkdir run
CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -o run/${prfx}_amd_linux
CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build -o run/${prfx}_amd_macos
CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build -o run/${prfx}_arm_macos
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o run/${prfx}_amd_windows

echo -n "" > embed/appl.txt
echo -n "" > embed/srvh.txt
echo -n "" > embed/svch.txt
echo -n "" > embed/cacr.txt
echo -n "" > embed/mycr.txt
echo -n "" > embed/pkey.txt
echo -n "" > embed/salt.txt
echo -n "" > embed/phrs.txt
echo -n "" > embed/type.txt
echo -n "" > embed/name.txt
echo -n "" > embed/vers.txt
echo -n "" > embed/hash.txt
echo -n "" > embed/manu.txt
echo -n "" > embed/desc.txt
echo -n "" > embed/envr.txt

export BIN_SRVH=$srvh
export BIN_SVCH=$svch
export BIN_HASH=$hash
export BIN_PKEY=$mypk
export BIN_CACR=$cacr
export BIN_MYCR=$mycr
export BIN_PHRS=$phrs
export BIN_SALT=$salt
export BIN_ENVR=$envr
export BIN_VERS=$vers
export BIN_TEAM=$team
export BIN_APIK=$apik
export BIN_BLOB_ACCT=$azbl
export BIN_BLOB_KEY=$azky
export BIN_GRAY=$gray

rm run/${whch}
envsubst < make/set.env > env.1p
op inject -i env.1p -o env.txt > /dev/null
cat env.txt >> run/${whch}
echo "" >> run/${whch}
echo "./${prfx}_${MYARCH}_${MYOS} \$@" >> run/${whch}
chmod ugo+x run/${whch}

#kubectl config use-context sss-svcs-${envr}-k8s

#if [ "$whch" == "service" ]; then
#  export PORT=$port
#  export APP=$nspc
#  cp ../make/service.yaml . > /dev/null 2>&1
#  envsubst < service.yaml | kubectl apply -f -
#  export svip=$(kubectl -n ${nspc} get services ${nspc}-service -o=jsonpath='{.status.loadBalancer.ingress[0].ip}')
#  echo "Setting service IP $svip in 1Pass"
#  op item edit "deployments" $envr.BIN_V5_HOST=$svip --vault=Configuration > /dev/null
#  serv="$(op read op://Configuration/deployments/${envr}/BIN_V5_HOST)"
#fi

rm -f cert-ext.txt cert.pem pkey.pem pkey.pem.b64 req.pem ca-*.*
rm -f deployment.yaml service.yaml Dockerfile Dockerfile.run
rm -f env.1p env.txt

cd make
