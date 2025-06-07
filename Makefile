SHELL=/usr/bin/env bash

check-executor:
	@if ! command -v executor &> /dev/null; then echo "Installing executor..."; brew tap github.com/jorgefuertes/executor; brew install executor; fi

gen: check-executor
	@executor run -d "Generating templates" -c "goht generate"

start-dev: check-executor
	@executor run -d "Starting Redis" -c "scripts/pod.sh redis start"

stop-dev: check-executor
	@executor run -d "Stopping Redis" -c "scripts/pod.sh redis stop"

status-dev: check-executor
	@executor run -d "Checking Redis" -c "scripts/pod.sh redis status"

test: start-dev gen
	@(set -e; err=0; \
		executor run -d "Running tests" -c "go test ./..." || err=$$?; \
		make stop-dev; \
		exit $$err)

test-clean: check-executor
	@executor run -d "Cleaning test cache" -c "go clean -testcache"
	@make test

lint: check-executor
	@executor run -d "staticcheck" -c "staticcheck ./..."
	@executor run -d "gofumpt" -c "gofumpt -d -l -extra ."
	@executor run -d "vet" -c "go vet ./..."
	@executor run -d "Linting with golangci-lint" -c "GOGC=80 /Users/queru/Desarrollo/gocode/bin/golangci-lint run --fast --concurrency 16"

run: start-dev gen
	go run cmd/testserver/main.go
	@make stop-dev
