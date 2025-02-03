# Project Name
BINARY_NAME=clx

# Go related variables.
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin
GOPKG=$(GOBASE)


# Go build and run commands
.PHONY: all build run clean format

all: build

format:
	@echo "📝 Formatting..."
	go fmt ./...

mod-tidy:
	@echo "🧹 Running go mod tidy..."
	@go mod tidy

build: format mod-tidy
	@echo "🚀 Building..."
	@GOBIN=$(GOBIN) go build -o $(GOBIN)/$(BINARY_NAME) $(GOPKG)

clean:
	@echo "🧹 Cleaning..."
	@GOBIN=$(GOBIN) go clean
	@rm -rf $(GOBIN)/*

test:
	@echo "🧪 Running Tests..."
	@go test $(TEST_FLAGS) -coverprofile=profile.cov ./...

release:
	@echo "🚀 Releasing..."
	goreleaser release --verbose --clean --timeout 90m