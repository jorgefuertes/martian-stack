SHELL=/usr/bin/env bash

gen:
	@executor run -d "Generating templates" -c "go tool goht generate"

start-dev:
	@executor run -d "Starting Redis" -c "scripts/pod.sh redis start"

stop-dev:
	@executor run -d "Stopping Redis" -c "scripts/pod.sh redis stop"

status-dev:
	@executor run -d "Checking Redis" -c "scripts/pod.sh redis status"

test: start-dev gen
	@(set -e; err=0; \
		executor run -d "Running tests" -c "go test ./..." || err=$$?; \
		make stop-dev; \
		exit $$err)

test-clean:
	@executor run -d "Cleaning test cache" -c "go clean -testcache"
	@make test

lint:
	@executor run -d "staticcheck" -c "go tool staticcheck ./..."
	@executor run -d "gofumpt" -c "go tool gofumpt -d -l -extra ."
	@executor run -d "vet" -c "go vet ./..."
	@executor run -d "golangci-lint" -c "GOGC=80 go tool golangci-lint run --fast --concurrency 16"
	@executor run -d "markdownlint" -c "trunk check --filter=markdownlint --all"
	@executor run -d "govulncheck" -c "go tool govulncheck ./..."

run: start-dev gen
	go run cmd/testserver/main.go
	@make stop-dev
