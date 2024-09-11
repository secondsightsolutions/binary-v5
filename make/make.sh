#!/bin/bash

MYOS=macos
MYARCH=arm

echo_usage() {
  echo "make.sh {build} {client|server} {branch|local} {staging|prod} {appl} {nspc} {descr} {name} {type}"
  echo "make.sh build   whch   from   envr    appl  nspc desc   name    type"
  #             1      2      3       4      5      6    7      8       9
  echo "make.sh build  client local  staging rebate rbt  'desc' teva    manu"
  echo "make.sh build  client master prod    rebate rbt  'desc' teva    manu"
  echo "make.sh build  client local  staging claims clm  'desc' util"
  echo "make.sh build  server local  staging claims clm  'desc' myImg"
  echo ""
  echo "make.sh {deploy} {client|server} {staging|prod} {appl} {nspc} {descr} {name}"
  echo "make.sh deploy whch   envr    appl      nspc   desc      name     readme"
  #             1      2      3       4         5      6         7        8
  echo "make.sh deploy client prod    binary-v4 bin   'version'  johnson_n_johnson    'readme.md"
  echo "make.sh deploy server staging rebate    rbt   'desc'     myImg"
  print_args
}
deploy_client() {
  local prfx=$1
  local name=$2
  local desc=$3
  local readme=$4
  repo="340B-ESP-Rebate-Binary-${name}"
  url="git@github.com:Second-Sight-Solutions/${repo}.git"
  rm -rf deploy
  mkdir deploy
  cd deploy
  git clone $url
  cd $repo
  git checkout main

  rm -rf 340B_ESP_modular_binary*
  cp ../../readme/${readme} README.md
  cp ../../${prfx}_amd_linux   340B_ESP_modular_binary_linux_amd
  cp ../../${prfx}_amd_windows 340B_ESP_modular_binary_windows_amd
  cp ../../${prfx}_amd_macos   340B_ESP_modular_binary_macos_amd
  cp ../../${prfx}_arm_macos   340B_ESP_modular_binary_macos_arm

  git add *
  git commit --amend -a -m "${desc}"
  git push --force

  cd ../..
  rm -rf deploy
}
goto_service_dir() {
  # Go into the directory of the service we're building or deploying.
  local repo=$1
  local whch=$2
  cpwd="$(pwd)"
  cd ..
}
gen_server_cert_and_pkey() {
  local appl=$1
  local envr=$2
  local whch=$3
  local phrs=$4
  local srvr=$5
  local name=$6
  local type=$7

  if [ -z "$name" ]; then
    name=${appl}
  fi
  if [ -z "$type" ]; then
    type="appl"
  fi

  echo "Generating server certificate and private key"

  rm -f cert-ext.txt ca-*.* cert.pem pkey.pem pkey.pem.b64 req.pem

  echo "$(op read ${all}/SSS_CA_CERT_PEM)" > ca-cert.pem
  echo "$(op read ${all}/SSS_CA_PKEY_PEM)" > ca-pkey.pem

  export X509_O=${envr}.secondsightsolutions.com
  export X509_OU=${name}.${type}.${whch}.${appl}
  export X509_CN=${name}
  export X509_EM=${name}@${type}.${whch}.${appl}.${envr}.secondsightsolutions.com

  local extcmd=""

  if [ "$whch" == "server" ]; then
    echo "Creating server certificate, embedding IP addresses"
    echo "subjectAltName=IP:0.0.0.0,IP:127.0.0.1,IP:${srvr}" > cert-ext.txt
    extcmd="-extfile cert-ext.txt"
  fi

  openssl req -newkey rsa:4096 -nodes -keyout pkey.pem -out req.pem -subj "/C=US/ST=DC/L=DC/O=$X509_O/OU=$X509_CN/CN=$X509_CN/emailAddress=$X509_EM" > /dev/null 2>&1
  openssl x509 -req -in req.pem -days 3650 -sha256 -CA ca-cert.pem -CAkey ca-pkey.pem -CAcreateserial -out cert.pem $extcmd > /dev/null 2>&1
  openssl verify -CAfile ca-cert.pem cert.pem
 # ../make/make --phrase=${phrs} --encrypt=pkey.pem --output=pkey.pem.b64
  ./make/crypt --phrase=${phrs} --encrypt=pkey.pem --output=pkey.pem.b64

  mypk="$(cat pkey.pem.b64)"
  mycr="$(base64 -i cert.pem)"
  cacr="$(base64 -i ca-cert.pem)"
}
create_service() {
    local nspc=$1
    local port=$2
    # Must create the service to get the public IP, must get added to the certificate.
    if [ "${envr}" == "dev" ]; then
    	kubectl config use-context sss-svcs-staging-k8s
    else
    	kubectl config use-context sss-svcs-${envr}-k8s
    fi
    export PORT=$port
    export APP=$nspc
    cp ../make/service.yaml . > /dev/null 2>&1
    envsubst < service.yaml | kubectl apply -f -
    export svip=$(kubectl -n ${nspc} get services ${nspc}-service -o=jsonpath='{.status.loadBalancer.ingress[0].ip}')
    echo "Setting service IP $svip in 1Pass"
    op item edit "deployments" $envr.BIN_GR_HOST=$svip --vault=Configuration > /dev/null
    serv="$(op read op://Configuration/deployments/${envr}/BIN_GR_HOST)"
}
write_embed_params() {
  echo -n $serv > embed/host.txt
  echo -n $port > embed/port.txt
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
    echo -n $name > embed/manu.txt
  fi
}
clear_embed_params() {
  echo -n "" > embed/host.txt
  echo -n "" > embed/port.txt
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
}
get_args() {
  if [ -z "$1" ]; then
    echo "Requires command as an argument"
    echo_usage
    exit 1
  fi

  if [ -z "$2" ]; then
    echo "Requires client or server as an argument"
    echo_usage
    exit 2
  fi

  # Set all the variables from the command line
  comd="$1"   # Build or deploy
  whch="$2"   # Server or client (or client-foo or client-bar, ...)
  if [ "$comd" == "build" ]; then
    brch="$3"   # Local or a branch name
    envr="$4"   # Dev, devint, staging, prod, etc.
    appl="$5"   # Proper applicatipon name like rebate or claims
    nspc="$6"   # Short name - namespace (like rbt or clm)
    desc="$7"   # Description - embedded in all apps, and in client/deploy (binary)
    name="$8"   # Identity, like brg or amgen
    type="$9"   # Interpreted by application (for rebate client, it's manufacturer or processor)

  elif [ "$comd" == "deploy" ]; then
    envr="$3"   # Dev, devint, staging, prod, etc.
    appl="$4"   # appl (rebate, claims)
    nspc="$5"   # namespace (rbt, clm)
    desc="$6"   # Description - embedded in app
    name="$7"   # Container name
    readme="$8" # Readme file
  fi

  # Validate the command line arguments
  if [ "$comd" == "build" ]; then
    if [ -z "$brch" ]; then
        echo "When building, must specify branch name or 'local'"
        echo_usage
        exit 3
    fi
    if [ -z "$envr" ]; then
      echo "Requires environment (dev, staging, prod, etc.) as an argument"
      echo_usage
      exit 4
    fi
    if [ -z "$appl" ]; then
      echo "Requires application name as an argument"
      echo_usage
      exit 5
    fi
    if [ -z "$nspc" ]; then
      echo "Requires short name (namespace) as an argument"
      echo_usage
      exit 6
    fi
    if [ -z "$desc" ]; then
      echo "Requires short description as an argument"
      echo_usage
      exit 7
    fi
    if [ -z "$type" ]; then
      echo "Defaulting to build type of proc"
      type="proc"
    fi
    # Server name is container tag, and is optional
    if [ "$whch" == "client" ]; then
      if [ -z "$name" ]; then
        echo "Requires application name as an argument"
        echo_usage
        exit 7
      fi
    fi
  elif [ "$comd" == "deploy" ]; then
    if [ -z "$envr" ]; then
        echo "Requires environment as an argument"
        echo_usage
        exit 9
    fi
    if [ -z "$appl" ]; then
        echo "Requires application as an argument"
        echo_usage
        exit 10
    fi
    if [ -z "$nspc" ]; then
        echo "Requires namespace as an argument"
        echo_usage
        exit 11
    fi
    if [ -z "$desc" ]; then
        echo "Requires short description as an argument"
        echo_usage
        exit 11
    fi
    if [ -z "$name" ]; then
        echo "Requires container tag as an argument"
        echo_usage
        exit 12
    fi
    if [ "$whch" == "client" ]; then
        if [ -z "$readme" ]; then
            echo "Requires readme file as an argument"
            echo_usage
            exit 13
        fi
    fi
  fi
}
get_data() {
  all="op://Configuration/all/all"
  cfg="op://Configuration/deployments/${envr}"

  repo="$appl"
  phrs="$(op read ${all}/SSS_PASSPHRASE)"
  salt="$(op read ${all}/SSS_SALT_ENCR_B64)"
  port="$(op read ${cfg}/BIN_GR_PORT)"
  vers="$(date '+%s')"
  hash=$(git rev-parse --short HEAD)
  team="$(op read ${cfg}/OG_TEAM)"
  apik="$(op read ${cfg}/OG_APIK)"
  azbl="$(op read ${cfg}/BIN_BLOB_ACCT)"
  azky="$(op read ${cfg}/BIN_BLOB_KEY)"
  gray="$(op read ${cfg}/GRAYLOG)"
}
write_runner_params() {
  local bin=$1
  echo "#!/bin/bash" > appl.run
  if [ -f "make/set.env" ]; then
    export BIN_HOST=$serv
    export BIN_PORT=$port
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
    
    envsubst < make/set.env > env.1p
    op inject -i env.1p -o env.txt > /dev/null
    cat env.txt >> appl.run
    echo "" >> appl.run
  fi

  echo "./${bin} \$@" >> appl.run
  echo ""
  echo "File content for appl.run"
  cat appl.run
  echo "--------------------------------"
  chmod ugo+x appl.run
}
write_secrets() {
  local nspc=$1
  export NSPC=$nspc
  export BIN_HOST=$serv
  export BIN_PORT=$port
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
  if [ -f "make/set.secrets" ]; then
    kubectl delete secret ${nspc}-secrets -n ${nspc}
    envsubst < make/set.secrets > set.secrets.run
    chmod +x set.secrets.run
    ./set.secrets.run
    rm set.secrets.run
  fi
}
print_args() {
  echo "Input arguments"
  echo "  comd=$comd"
  echo "  whch=$whch"
  echo "  brch=$brch"
  echo "  envr=$envr"
  echo "  appl=$appl"
  echo "  nspc=$nspc"
  echo "  name=$name"
  echo "  type=$type"
  echo "  desc=$desc"
  echo "  readme=$readme"
  echo ""
  echo "1Pass/derived arguments"
  echo "  repo=$repo"
  echo "  phrs=$phrs"
  echo "  salt=$salt"
  echo "  serv=$serv"
  echo "  port=$port"
  echo "  vers=$vers"
  echo "  hash=$hash"
  echo ""
}
build_full() {
  local prfx=$1
  CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -o ${prfx}_amd_linux
  CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build -o ${prfx}_amd_macos
  CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build -o ${prfx}_arm_macos
  CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o ${prfx}_amd_windows
}
build_stripped() {
  local prfx=$1
  # Example using garble. Note that garbled apps seem to run almost half speed!
  #CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 garble -literals -tiny build -ldflags="-s -w" -o ${prfx}_amd_linux

  CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -o ${prfx}_amd_linux   -ldflags="-s -w"
  CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build -o ${prfx}_amd_macos   -ldflags="-s -w"
  CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build -o ${prfx}_arm_macos   -ldflags="-s -w"
  CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o ${prfx}_amd_windows -ldflags="-s -w"
}
build_container() {
  APP=${bin}_amd_linux
  PORT=${port}
  TAG=$( tr -s ' ' '_' <<< "$appl $brch $name $hash $vers")
  IMAG="sssallimages.azurecr.io/${nspc}:${TAG}"

  az acr login -n sssallimages
  cp make/Dockerfile .
  envsubst < Dockerfile > Dockerfile.run
  docker build . -f Dockerfile.run -t ${IMAG}
  docker push ${IMAG}
  echo "Image tag is ${TAG}"
}

get_args $1 $2 $3 $4 $5 "$6" "$7" "$8" "$9"
get_data
goto_service_dir $repo $whch
#create_service $nspc $port
print_args
serv=127.0.0.1
gen_server_cert_and_pkey $appl $envr $whch $phrs $serv $name $type

if [ "$comd" == "build" ]; then
  bin="${appl}_${whch}"
  echo "Running build on ${bin}"
  export PATH=$GOPATH/bin:$PATH
  write_embed_params
  write_runner_params ${bin}_${MYARCH}_${MYOS}

  if [ "$brch" != "local" ]; then
    git stash -a
    git checkout $brch
  fi

  if [ "$whch" == "client" ]; then
    if [ "$envr" == "prod" ]; then
      build_stripped ${bin}
    else
      build_full ${bin}
    fi

  elif [ "$whch" == "server" ]; then
    build_full ${bin}
    if [ ! -z "$name" ]; then
      build_container
    fi
  else
    echo "WHICH $which not recognized"
    exit 13
  fi

  if [ "$whch" == "client" ]; then
    echo ""
    echo "Application version information"
    ./${bin}_${MYARCH}_${MYOS} -version  # Hopefully everything we build has this option!
    echo "--------------------------------"
  fi

  clear_embed_params
  echo ""
  echo "Build complete"

  if [ "$brch" != "local" ]; then
    echo "Restoring stash"
    git stash pop
    git stash clear
  fi

elif [ "$comd" == "deploy" ]; then
  bin="${appl}_${whch}"
  if [ "$whch" == "client" ]; then
      if [ "$envr" == "staging" ]; then
        echo "deploy client to staging"
        # TODO: add staging upload
      else
        echo "deploy client to production"
        deploy_client ${bin} ${name} ${desc} ${readme}
      fi
  elif [ "$whch" == "server" ]; then
      echo "deploy server"
      if [ "${envr}" == "dev" ]; then
      	kubectl config use-context sss-svcs-staging-k8s
      else
      	kubectl config use-context sss-svcs-${envr}-k8s
      fi
      write_secrets $nspc
      export TAG=${name}
      export PORT=${port}
      export ACR=sssallimages.azurecr.io
      export SECR="${nspc}-secrets"
      export APP=${nspc}
      if [ "$envr" == "staging" ]; then
        export REPL="3"
      else
        export REPL="3"
      fi
      #cp ../../grpc-services/make/deployment.yaml .
      envsubst < make/deployment.yaml | kubectl apply -f -
  else
      echo "WHICH $whch not recognized"
      exit 14
  fi
else
  echo "Command $comd not recognized"
  exit 15
fi

rm -f cert-ext.txt cert.pem pkey.pem pkey.pem.b64 req.pem ca-*.*
rm -f deployment.yaml service.yaml Dockerfile Dockerfile.run
rm -f env.1p env.txt

cd $cpwd
