SHELL := /bin/sh
GOPATH ?= $(shell go env GOPATH)
SWAG ?= $(GOPATH)/bin/swag

.PHONY: run test migrate compose-up compose-up-db compose-up-app compose-down sqlc swagger

TEST_PACKAGES ?= ./...

compose-up:
	docker compose up -d mysql app

compose-up-db:
	docker compose up -d mysql

compose-up-app:
	docker compose up -d app

compose-down:
	docker compose down -v

migrate:
	go run ./cmd/migrate

run: compose-up-db migrate compose-up-app

sqlc:
	sqlc generate

swagger:
	$(SWAG) init -g cmd/api/main.go -o docs

test:
	go test $(TEST_PACKAGES)
