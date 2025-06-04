BINARY_NAME=alacritty-colors
BUILD_DIR=build

.PHONY: build clean install test run

build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) cmd/alacritty-colors/main.go

clean:
	rm -rf $(BUILD_DIR)

install: build
	cp $(BUILD_DIR)/$(BINARY_NAME) ~/bin/

test:
	go test ./...

run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

deps:
	go mod tidy