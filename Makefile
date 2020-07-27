GIT_CURRENT_BRANCH := ${shell git symbolic-ref --short HEAD}

.PHONY: help clean test run

.DEFAULT: help

help:
	@echo "make docker:"
	@echo "       Run app with docker and docker-compose"
	@echo ""
	@echo "make dockerdown:"
	@echo "       Remove app from docker with docker-compose down"
	@echo ""

docker:
	@echo "---- Building & Up Container ----"
	@docker-compose -f compose-kafka.yml down
	@docker-compose -f compose-kafka.yml up -d
	@sleep 3
	@docker-compose down
	@docker-compose build	
	@docker-compose up -d
	@sleep 3
	@echo "---- EndPoints ----"
	@echo "---- Consumer API - http://127.0.0.1:8084/index ----"

dockerdown:
	@docker-compose down
	@docker-compose -f compose-kafka.yml down

run:
	@export BROKER_SRV=localhost
	@export BROKER_PORT=9092
	@export TOPIC=events
	@go run producer/main.go