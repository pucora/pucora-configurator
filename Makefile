.PHONY: build test install clean presets

BINARY := velonetics-config
BUILD_DIR := ./bin

build:
	go build -o $(BUILD_DIR)/$(BINARY) ./cmd/velonetics-config

install:
	go install ./cmd/velonetics-config

test:
	go test ./...

clean:
	rm -rf $(BUILD_DIR) output

# Regenerate embedded presets from profiles/
presets:
	cp profiles/*.yaml internal/presets/profiles/

# Example: generate config from rest-proxy preset
example:
	$(BUILD_DIR)/$(BINARY) presets apply rest-proxy -g ./output
