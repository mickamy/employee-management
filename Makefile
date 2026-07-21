BUILD_DIR = bin

.PHONY: build clean test lint

build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/server ./cmd/server

clean:
	@echo "Cleaning up..."
	rm -rf $(BUILD_DIR)

test:
	go test ./... -race

lint:
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "golangci-lint is not installed"; \
		exit 1; \
	}
	golangci-lint run
