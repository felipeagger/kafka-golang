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

build:
	@docker build -t kafkago-producer:latest -f ./Dockerfile-Producer .
	@docker build -t kafkago-consumer:latest -f ./Dockerfile-Consumer .

docker:
	@echo "---- Building & Up Container ----"
	@docker-compose down
	@docker-compose up -d --build

dockerdown:
	@docker-compose down
	@docker-compose -f compose-kafka.yml down

run:
	@echo "Run: export BROKERS=localhost:1909,localhost:2909"
	@echo "Run: export TOPIC=events"
	@echo "Run: export GROUP=Consumers"
	@go run producer/*.go & go run consumer/*.go