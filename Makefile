SHELL := /bin/bash -euxo pipefail

NAME := envtab

VERSION != git describe --tags --always --dirty | sed 's/^v//'
COMMIT != git rev-parse --short HEAD
BUILD_DATE != date -u +"%Y-%m-%dT%H:%M:%SZ"

# LDFLAGS for version injection (see cmd/root.go)
LDFLAGS := -X 'github.com/gmherb/envtab/cmd.Version=$(VERSION)' \
           -X 'github.com/gmherb/envtab/cmd.Commit=$(COMMIT)' \
           -X 'github.com/gmherb/envtab/cmd.BuildDate=$(BUILD_DATE)'

.PHONY: all
all: test install docs

.PHONY: build
build:
	@echo "Building $(NAME) version $(VERSION) (commit $(COMMIT))"
	@go build -ldflags "$(LDFLAGS)" -o $(NAME)
	@chmod +x $(NAME)

.PHONY: install
install: build
	@echo "Installing $(NAME) version $(VERSION)"
	@[[ $(shell id -u) == 0 ]] \
		&& mv $(NAME) /usr/local/bin \
		|| sudo mv $(NAME) /usr/local/bin

.PHONY: version
version:
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Build Date: $(BUILD_DATE)"

.PHONY: docs
docs:
	@rm -f docs/*.md
	@go run ./tools/gen-docs.go

.PHONY: test
test:
	@go mod tidy
	@go mod verify
	@rm -f profile.cov
	@go clean -testcache
	@export PATH="$$HOME/go/bin:$$PATH" && go test -v -p 1 -timeout 30s -coverpkg ./... -coverprofile=profile.cov ./...
	@test -f profile.cov && go tool cover -func profile.cov
