.PHONY: build run install bootstrap-user test fmt fmt-check up down

build:
	go build -o bin/archon-api ./cmd/archon-api
	go build -o bin/archon ./cmd/archon

run:
	go run ./cmd/archon-api

install:
	go run ./cmd/archon install $(ARGS)

bootstrap-user:
	go run ./cmd/archon bootstrap-user $(ARGS)

test:
	go test ./...

fmt:
	gofmt -w $$(find . -type f -name '*.go' -not -path './vendor/*')

fmt-check:
	test -z "$$(gofmt -l $$(find . -type f -name '*.go' -not -path './vendor/*'))"

up:
	docker compose -f deploy/docker-compose.yml up --build

down:
	docker compose -f deploy/docker-compose.yml down
