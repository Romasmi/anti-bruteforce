ifneq (,$(wildcard deployments/.env))
	include deployments/.env
	export
endif

BIN := "./bin/anti-bruteforce"
DOCKER_IMG="anti-bruteforce:develop"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X 'main.release=develop' -X 'main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S)' -X 'main.gitHash=$(GIT_HASH)'

#Postrges
POSTGRES_USER ?= postgres
POSTGRES_PASSWORD ?= password
POSTGRES_DB ?= backend
POSTGRES_PORT ?= 5435
POSTGRES_CONTAINER := postgres-anti-bruteforce

up:
	docker compose -f deployments/docker-compose.yaml up -d --build

down:
	docker compose -f deployments/docker-compose.yaml down

build:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd/anti-bruteforce

run: build
	$(BIN) -config ./configs/config.yaml

build-img:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG) \
		-f deployments/api.Dockerfile .

run-img: build-img
	docker run $(DOCKER_IMG)

version: build
	$(BIN) version

test:
	go test -v -race ./internal/...

lint:
	go mod download
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.5 run ./...

migration:
	go run github.com/pressly/goose/v3/cmd/goose@latest -dir migrations create $(NAME) sql

generate:
	mkdir -p pkg/api
	protoc --proto_path=api \
		--proto_path=internal/proto \
		--go_out=pkg/api --go_opt=paths=source_relative \
		--go-grpc_out=pkg/api --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=pkg/api --grpc-gateway_opt=paths=source_relative \
		--openapiv2_out=pkg/api --openapiv2_opt=logtostderr=true \
		api/AntiBruteforce.proto

puml:
	plantuml -tsvg pkg/ratelimiter/docs/*.puml

#load-tests:
#	k6 run load_tests/<add test later>

integration-tests:
	docker compose -f deployments/docker-compose.integration.yaml up --build --exit-code-from integration-tests; \
	RET=$$?; \
	docker compose -f deployments/docker-compose.integration.yaml down; \
	exit $$RET

clean:
	rm -rf bin/
	go clean -cache -testcache

.PHONY: build run run-api run-all build-img run-img version test api-test unit-test lint migration generate puml load-tests up down integration-tests clean
