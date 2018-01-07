export GITHUB_CLIENT_ID=replace-with-your-github-oauth-client-id
export GITHUB_CLIENT_SECRET=replace-with-your-github-oauth-client-secret
export GITHUB_WEBHOOK_SECRET=replace-with-your-github-webhook-secret
export SESSION_SECRET=change-this-to-any-random-string

export DOCKER_IMAGE_REPO=xovox #dockerhub repository where the test container images is hosted
export PORT=8080
export ROOT_DIR=$(shell pwd)

all: restart

start: start_ci start_app	

stop: stop_ci

restart: stop start

start_app:
	@go run ${ROOT_DIR}/app/*.go
	
start_ci:
	# startup ci resources
	@echo "Setting up resources"
	@docker-compose -f ci/docker-compose.yml up -d 

stop_ci:
	# shutdown ci resources
	@echo "Stopping resources"
	@docker-compose -f ci/docker-compose.yml stop