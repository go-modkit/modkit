SHELL := /bin/sh
GOPATH ?= $(shell go env GOPATH)
SWAG ?= $(GOPATH)/bin/swag

.PHONY: run test migrate seed compose-up compose-up-db compose-up-app compose-down sqlc swagger

TEST_PACKAGES ?= ./...

compose-up:
	docker compose up -d --build mysql app

compose-up-db:
	docker compose up -d mysql

compose-up-app:
	docker compose up -d --build app

compose-down:
	docker compose down -v

migrate:
	go run ./cmd/migrate

seed:
	go run ./cmd/seed

run: compose-up-db migrate seed compose-up-app

sqlc:
	sqlc generate

swagger:
	$(SWAG) init -g cmd/api/main.go -o docs

test:
	go test $(TEST_PACKAGES)
