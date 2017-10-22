export ROOT_DIR=$(shell pwd)
export DOCKER_IMAGE_REPO=xovox
export PORT=8080

all: restart

start: start_ci start_app	

stop: stop_ci

restart: stop start

start_app:
	@go run ./app/*.go
	
start_ci:
	# startup ci resources
	@echo "Setting up resources"
	@docker-compose -f ci/docker-compose.yml up -d 

stop_ci:
	# shutdown ci resources
	@echo "Stopping resources"
	@docker-compose -f ci/docker-compose.yml stop