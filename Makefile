
PHONY: test
test:
	@go test -v -coverpkg ./... -coverprofile=profile.cov ./...
	@go tool cover -func profile.cov