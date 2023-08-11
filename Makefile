.PHONY: build
build:
	@goreleaser build --snapshot --rm-dist

.PHONY: test
test:
	@go test ./...

.PHONY: release
release:
	@goreleaser release --rm-dist
