.PHONY: build run test fmt up down

build:
	go build -o bin/archon-api ./cmd/archon-api
	go build -o bin/archon ./cmd/archon

run:
	go run ./cmd/archon-api

test:
	go test ./...

fmt:
	gofmt -w .

up:
	docker compose -f deploy/docker-compose.yml up --build

down:
	docker compose -f deploy/docker-compose.yml down
