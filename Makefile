SHELL := /bin/bash -euxo pipefail

NAME := envtab

.PHONY: test
test:
	@go mod tidy
	@go mod verify
	@PATH="$$HOME/go/bin:$$PATH" go test -v -p 1 -timeout 30s -coverpkg ./... -coverprofile=profile.cov ./...
	@go tool cover -func profile.cov

.PHONY: build
build:
	@go build -o $(NAME)
	@chmod +x $(NAME)

.PHONY: docs
docs:
	@rm -f docs/*.md
	@go run ./tools/gen-docs.go

.PHONY: install
install: build
	@[[ $(shell id -u) == 0 ]] \
		&& mv $(NAME) /usr/local/bin \
		|| sudo mv $(NAME) /usr/local/bin

.PHONY: all
all: test build docs install
