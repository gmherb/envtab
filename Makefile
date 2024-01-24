SHELL := /bin/bash -euxo pipefail

NAME := envtab

.PHONY: test
test:
	@go test -v -coverpkg ./... -coverprofile=profile.cov ./...
	@go tool cover -func profile.cov

.PHONY: build
build:
	@go build -o $(NAME)
	@chmod +x $(NAME)

.PHONY: install
install:
	@[[ $(shell id -u) == 0 ]] \
		&& mv $(NAME) /usr/local/bin \
		|| sudo mv $(NAME) /usr/local/bin
