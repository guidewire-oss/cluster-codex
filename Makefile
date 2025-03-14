# Project Name
BINARY_NAME=clx

# Go related variables.
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin
GOPKG=$(GOBASE)
# You can use boolean logic in the LABEL_FILTER, eg export LABEL_FILTER="integration || unit"
export LABEL_FILTER ?= unit

# Go build and run commands
.PHONY: all build run clean format test unit-test integration-test setup-test-data teardown-test-data

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

setup-test-data:
	@echo "🚀 Applying Kubernetes Test Data..."
	kubectl apply -f test/k8s/test-data.yaml

teardown-test-data:
	@echo "🧹 Cleaning Up Kubernetes Test Data..."
	kubectl delete -f test/k8s/test-data.yaml || true

unit-test:
	@echo "🧪 Running Unit Tests..."
	ginkgo -r -p --label-filter=unit --randomize-all

test:
	@echo "🧪 Running All Tests with labels \"$(LABEL_FILTER)\"..."
	ginkgo -r -p --label-filter="$(LABEL_FILTER)" --succinct --randomize-all

# Use this target in the workflow since it handles setting up and tearing down the test data.
int-test:
	@echo "🧪 Running Integration Tests..."
	ginkgo -r -p --label-filter=integration --succinct --randomize-all

# Use this target when running locally.
integration-test: setup-test-data
	# Ensure that the test data is torn down locally when there is a test failure
	@$(MAKE) int-test || true
	@$(MAKE) teardown-test-data

release:
	@echo "🚀 Releasing..."
	goreleaser release --verbose --clean --timeout 90m
