BINARY_NAME=gotelemetry
MAIN_PATH=./cmd/gotelemetry
BUILD_ENV=CGO_ENABLED=0 GOOS=linux
VERSION=0.6-core

# You can hardcode the CONFIG_PATH if you don't want to use env's
CONFIG_PATH ?= 

LDFLAGS=-s -w -X main.BuildVersion=$(VERSION)

ifneq ($(CONFIG_PATH),)
LDFLAGS += -X main.ConfigPath=$(CONFIG_PATH)
endif

.PHONY: all build run clean install

all: build

build:
	@echo "==> Compiling $(BINARY_NAME)..."
	@mkdir -p bin
	$(BUILD_ENV) go build -ldflags="$(LDFLAGS)" -o bin/$(BINARY_NAME) $(MAIN_PATH)

install: build
	@echo "==> Installing binary in /usr/local/bin..."
	@sudo cp bin/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	@sudo chmod +x /usr/local/bin/$(BINARY_NAME)

clean:
	@rm -rf bin/
