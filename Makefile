.PHONY: build test install clean presets dev-api dev-web build-api build-web build-all docker-up

BINARY := pucora-config
API_BINARY := pucora-config-api
BUILD_DIR := ./bin

build:
	go build -o $(BUILD_DIR)/$(BINARY) ./cmd/pucora-config

build-api:
	go build -o $(BUILD_DIR)/$(API_BINARY) ./cmd/pucora-config-api

build-web:
	cd web && npm run build

build-all: build build-api build-web

install:
	go install ./cmd/pucora-config
	go install ./cmd/pucora-config-api

dev-api:
	go run ./cmd/pucora-config-api

dev-web:
	cd web && npm run dev

docker-up:
	docker compose -f deploy/docker-compose.yml up --build

test:
	go test ./cmd/... ./internal/...

clean:
	rm -rf $(BUILD_DIR) output output-* web/dist

# Regenerate embedded presets from profiles/
presets:
	cp profiles/*.yaml internal/presets/profiles/

# Example: generate config from rest-proxy preset
example:
	$(BUILD_DIR)/$(BINARY) presets apply rest-proxy -g ./output
