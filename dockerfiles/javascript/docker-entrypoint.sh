#!/bin/bash
trap 'exit' ERR

source /etc/profile

echo "<h3>Starting the build</h3>"

echo "<h3>Adding SSH keys</h3>"
mkdir -p /root/.ssh/ && cp -R .ssh/* "$_"
chmod 600 /root/.ssh/* &&\
    ssh-keyscan github.com > /root/.ssh/known_hosts &&\
    ssh-keyscan bitbucket.com >> /root/.ssh/known_hosts
echo 

echo "<h3>Checkout source code</h3>"
git clone ${PROJECT_REPOSITORY_URL} ${PROJECT_REPOSITORY_NAME} 
cd ${PROJECT_REPOSITORY_NAME}
git checkout ${PROJECT_BRANCH}
echo

# check if sicuro.json is present
SICURO_CONFIG_PRESENT=false
SICURO_CONFIG_FILE=./sicuro.json

if [ -r ./sicuro.json ]; then
    SICURO_CONFIG_PRESENT=true
fi

echo "<h3>Dependencies</h3>"
if ! ($SICURO_CONFIG_PRESENT && $(cat $SICURO_CONFIG_FILE | jq --raw-output '. | .dependencies.override//false'))  ; then
    # default language dependencies
    echo Exporting NODE_ENV
    export NODE_ENV=test
fi
if $SICURO_CONFIG_PRESENT ; then
    source <(cat $SICURO_CONFIG_FILE | jq --raw-output '. | .dependencies.custom[]?')
fi

echo "<h3>Setup</h3>"
if ! ($SICURO_CONFIG_PRESENT && $(cat $SICURO_CONFIG_FILE | jq --raw-output '. | .setup.override//false')); then
    # default language setup
    npm install
fi
if $SICURO_CONFIG_PRESENT ; then
    source <(cat $SICURO_CONFIG_FILE | jq --raw-output '. | .setup.custom[]?')
fi

echo "<h3>Test</h3>"
if ! ($SICURO_CONFIG_PRESENT && $(cat $SICURO_CONFIG_FILE | jq --raw-output '. | .test.override//false')); then
    # default language test command
    npm test
fi
if $SICURO_CONFIG_PRESENT ; then
    source <(cat $SICURO_CONFIG_FILE | jq --raw-output '. | .test.custom[]?')
fi

exec "$@"
