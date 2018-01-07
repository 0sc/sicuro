#!/bin/bash
trap 'exit' ERR

# the first argument, $1, is the a list of the enviroment variables
# mostly those for resources such as 
# DATABASE_URL, REDIS_URL etc
# example: -e DATABASE_URL=postgres://postgres@postgres:5432 -e REDIS_URL=redis://redis
# ----
# the second argument is the language of the project
# which is gotten from the github details for the project

DOCKER_ENVS=${1}
DOCKER_IMAGE=${2}

docker run --rm -v ${CI_DIR}/.ssh:/.ssh \
				--network ci_default $DOCKER_ENVS $DOCKER_IMAGE
				