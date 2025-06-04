BINARY_NAME=alacritty-colors
BUILD_DIR=build

.PHONY: build clean install test run

build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) cmd/alacritty-colors/main.go

clean:
	rm -rf $(BUILD_DIR)

install: build
	@echo "Installing to /usr/local/bin/ (requires sudo)..."
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/

# Install to user's local bin directory (no sudo required)
local-install: build
	@mkdir -p ~/.local/bin
	cp $(BUILD_DIR)/$(BINARY_NAME) ~/.local/bin/
	@echo "Installed to ~/.local/bin/"
	@echo "Make sure ~/.local/bin is in your PATH"

test:
	go test ./...

run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

deps:
	go mod tidy