.PHONY: test test-unit test-integration

test: test-unit test-integration

# Run fast unit tests (default build tags)
test-unit:
	go test ./...

# Run integration-only tests marked with the integration tag
test-integration:
	go test -tags=integration ./...
