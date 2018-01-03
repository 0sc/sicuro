#!/bin/bash
trap 'exit' ERR

source /etc/profile
# [[ $- == *i* ]] && echo 'Interactive' || echo 'Not interactive'
type rvm | head -1
echo "<h3>Starting the build</h3>"

echo "<h3>Adding SSH keys</h3>"
mkdir -p /root/.ssh/ && cp -R .ssh/* "$_"
chmod 400 /root/.ssh/* &&\
    ssh-keyscan github.com > /root/.ssh/known_hosts &&\
    ssh-keyscan bitbucket.com >> /root/.ssh/known_hosts
echo 

echo "<h3>Checkout source code</h3>"
git clone ${PROJECT_REPOSITORY_URL} ${PROJECT_REPOSITORY_NAME} 
cd ${PROJECT_REPOSITORY_NAME}
git checkout ${PROJECT_BRANCH}
# Have rvm recheck ruby version
cd .
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
    echo Exporting RAILS_ENV
    export RAILS_ENV=test
    echo Exporting RACK_ENV
    export RACK_ENV=test
    echo Exporting SECRET_KEY_BASE
    export SECRET_KEY_BASE=aec45e599f914e0547504f7013057c6af37843d7003cc642d775eb8af4fb2b0101faf1c38470f1a30348787f5782918741b116262af2d54
    Xvfb :99 & export DISPLAY=:99
fi
if $SICURO_CONFIG_PRESENT ; then
    source <(cat $SICURO_CONFIG_FILE | jq --raw-output '. | .dependencies.custom[]?')
fi

echo "<h3>Setup</h3>"
if ! ($SICURO_CONFIG_PRESENT && $(cat $SICURO_CONFIG_FILE | jq --raw-output '. | .setup.override//false')); then
    # default language setup
    bundle install
    bundle exec rake db:create db:schema:load --trace
fi
if $SICURO_CONFIG_PRESENT ; then
    source <(cat $SICURO_CONFIG_FILE | jq --raw-output '. | .setup.custom[]?')
fi

echo "<h3>Test</h3>"
if ! ($SICURO_CONFIG_PRESENT && $(cat $SICURO_CONFIG_FILE | jq --raw-output '. | .test.override//false')); then
    # default language test
    bundle exec rake test
fi
if $SICURO_CONFIG_PRESENT ; then
    source <(cat $SICURO_CONFIG_FILE | jq --raw-output '. | .test.custom[]?')
fi

exec "$@"
