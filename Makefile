.PHONY: build clean install install-local uninstall test run help

APP_NAME := tide
BUILD_DIR := build
GO := go
PREFIX ?= /usr/local

# Build flags
LDFLAGS := -s -w

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-16s\033[0m %s\n", $$1, $$2}'

build: ## Build for current platform
	$(GO) build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME) .

build-linux: ## Build for Linux amd64
	GOOS=linux GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 .

build-darwin: ## Build for macOS amd64
	GOOS=darwin GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 .

build-darwin-arm: ## Build for macOS arm64 (Apple Silicon)
	GOOS=darwin GOARCH=arm64 $(GO) build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-darwin-arm64 .

build-all: build-linux build-darwin build-darwin-arm ## Build for all platforms

install: build ## Install to /usr/local/bin (requires sudo)
	@echo "Installing tide to $(PREFIX)/bin ..."
	@if [ -w "$(PREFIX)/bin" ]; then \
		install -m 755 $(BUILD_DIR)/$(APP_NAME) $(PREFIX)/bin/$(APP_NAME); \
	else \
		sudo install -m 755 $(BUILD_DIR)/$(APP_NAME) $(PREFIX)/bin/$(APP_NAME); \
	fi
	@echo "✓ tide installed. Run 'tide --help' to get started."

install-local: build ## Install to ~/.local/bin (no sudo)
	@mkdir -p $(HOME)/.local/bin
	install -m 755 $(BUILD_DIR)/$(APP_NAME) $(HOME)/.local/bin/$(APP_NAME)
	@echo "✓ tide installed to ~/.local/bin"
	@echo "  Make sure ~/.local/bin is in your PATH:"
	@echo "  export PATH=\"$(HOME)/.local/bin:\$$PATH\""

uninstall: ## Remove tide from system
	@for dir in $(PREFIX)/bin $(HOME)/.local/bin; do \
		if [ -f "$$dir/$(APP_NAME)" ]; then \
			rm -f "$$dir/$(APP_NAME)"; \
			echo "Removed $$dir/$(APP_NAME)"; \
		fi \
	done

test: ## Run tests
	$(GO) test ./... -v -count=1

test-race: ## Run tests with race detector
	$(GO) test ./... -v -race -count=1

clean: ## Remove build artifacts
	rm -rf $(BUILD_DIR)

run: build ## Build and run
	./$(BUILD_DIR)/$(APP_NAME) list

fmt: ## Format code
	$(GO) fmt ./...

vet: ## Run go vet
	$(GO) vet ./...

lint: fmt vet ## Lint and format
