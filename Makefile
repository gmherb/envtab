SHELL := /bin/bash -euxo pipefail

NAME := envtab

.PHONY: test
test:
	@PATH="$$HOME/go/bin:$$PATH" go test -v -p 1 -timeout 30s -coverpkg ./... -coverprofile=profile.cov ./...
	@go tool cover -func profile.cov

.PHONY: build
build: docs
	@go build -o $(NAME)
	@chmod +x $(NAME)

docs:
	@go run ./tools/gen-docs.go

.PHONY: install
install:
	@[[ $(shell id -u) == 0 ]] \
		&& mv $(NAME) /usr/local/bin \
		|| sudo mv $(NAME) /usr/local/bin

.PHONY: all
all: test build docs install
