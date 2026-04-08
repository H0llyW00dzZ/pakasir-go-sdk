# Copyright 2026 H0llyW00dzZ
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

.PHONY: all proto lint-proto build test test-cover test-e2e vet fmt gocyclo clean deps header

# Test packages (excludes generated proto code that has no tests).
TEST_PKGS := $(shell go list ./src/... | grep -v /grpc/pakasir/)

# Print ASCII art banner.
header:
	@printf '%s\n' '  ____        _               _        '
	@printf '%s\n' ' |  _ \ __ _ | | __ __ _ ___ (_) _ __  '
	@printf '%s\n' ' | |_) / _` || |/ // _` / __|| || '"'"'__| '
	@printf '%s\n' ' |  __/ (_| ||   <| (_| \__ \| || |    '
	@printf '%s\n' ' |_|   \__,_||_|\_\\__,_|___/|_||_|    '
	@printf '%s\n' '                                        '
	@printf '%s\n' '  Go SDK by H0llyW00dzZ (@github.com/H0llyW00dzZ)'
	@echo ""

# Default target.
all: header build test

## ──────────────────────────────────────────────
## Proto Generation
## ──────────────────────────────────────────────

# Generate Go code from proto files using buf.
proto: header
	@echo "==> Generating proto..."
	buf generate
	@echo "==> Done."

# Lint proto files.
lint-proto: header
	@echo "==> Linting proto files..."
	buf lint
	@echo "==> Done."

## ──────────────────────────────────────────────
## Build
## ──────────────────────────────────────────────

# Build all packages (compile check, no binary output).
build: header
	@echo "==> Building all packages..."
	go build ./...
	@echo "==> Done."

## ──────────────────────────────────────────────
## Quality
## ──────────────────────────────────────────────

# Run all tests with race detector.
test: header
	@echo "==> Running tests..."
	go test $(TEST_PKGS) -race -v -count=1
	@echo "==> Done."

# Run tests with coverage report.
# To view in browser: go tool cover -html=coverage.txt
test-cover: header
	@echo "==> Running tests with coverage..."
	go test $(TEST_PKGS) -race -v -coverprofile=coverage.txt -covermode=atomic -count=1
	go tool cover -func=coverage.txt
	@echo "==> Done. (To view in browser: go tool cover -html=coverage.txt)"

# Run the gRPC end-to-end payment flow test.
# This exercises the full lifecycle: create -> simulate pay -> verify completed.
test-e2e: header
	@echo "==> Running gRPC E2E test..."
	go test -v -race -run TestE2EPaymentFlowSuccess ./src/grpc/
	@echo "==> Done."

# Run go vet.
vet: header
	@echo "==> Running go vet..."
	go vet ./src/...
	@echo "==> Done."

# Run cyclomatic complexity analysis (requires gocyclo).
# Scans src/ only — generated protobuf code in grpc/pakasir/v1 is excluded.
# Usage:
#   make gocyclo                    # report functions with complexity > 10
#   make gocyclo CYCLO_THRESHOLD=15 # custom threshold
CYCLO_THRESHOLD ?= 10
gocyclo: header
	@echo "==> Running gocyclo (threshold=$(CYCLO_THRESHOLD))..."
	gocyclo -over $(CYCLO_THRESHOLD) $(shell go list -f '{{.Dir}}' ./src/... | grep -v /grpc/pakasir/)
	@echo "==> Done."

# Check formatting (must produce no output).
fmt: header
	@echo "==> Checking gofmt..."
	@DIFF=$$(gofmt -s -d .); \
	if [ -n "$$DIFF" ]; then \
		echo "$$DIFF"; \
		echo ""; \
		echo "ERROR: gofmt found formatting issues. Run: gofmt -s -w ."; \
		exit 1; \
	fi
	@echo "==> Done."

## ──────────────────────────────────────────────
## Cleanup
## ──────────────────────────────────────────────

# Remove generated coverage files.
clean: header
	rm -f coverage.txt
	@echo "==> Cleaned."

## ──────────────────────────────────────────────
## Dependencies
## ──────────────────────────────────────────────

# Install required tools for proto generation and analysis.
deps: header
	go install github.com/bufbuild/buf/cmd/buf@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
	@echo "==> Done."
