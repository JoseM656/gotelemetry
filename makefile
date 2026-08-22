BINARY_NAME=gotelemetry
MAIN_PATH=./cmd/gotelemetry
BUILD_ENV=CGO_ENABLED=0 GOOS=linux
VERSION=0.5

.PHONY: all build run clean install

all: build

build:
	@echo "==> Compiling $(BINARY_NAME)..."
	@mkdir -p bin
	$(BUILD_ENV) go build -ldflags="-s -w" -o bin/$(BINARY_NAME) $(MAIN_PATH)

install: build
	@echo "==> Installing binary in /usr/local/bin..."
	@sudo cp bin/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	@sudo chmod +x /usr/local/bin/$(BINARY_NAME)

clean:
	@rm -rf bin/
