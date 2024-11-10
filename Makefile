SHELL=/usr/bin/env bash

check-executor:
	@if ! command -v executor &> /dev/null; then echo "Installing executor..."; brew tap github.com/jorgefuertes/executor; brew install executor; fi

gen: check-executor
	@executor run -d "Generating templates" -c "templ generate -lazy"

start-dev: check-executor
	@executor run -d "Starting MongoDB" -c "scripts/pod.sh mongo start"
	@executor run -d "Starting Redis" -c "scripts/pod.sh redis start"

stop-dev: check-executor
	@executor run -d "Stopping MongoDB" -c "scripts/pod.sh mongo stop"
	@executor run -d "Stopping Redis" -c "scripts/pod.sh redis stop"

status-dev: check-executor
	@executor run -d "Checking MongoDB" -c "scripts/pod.sh mongo status"
	@executor run -d "Checking Redis" -c "scripts/pod.sh redis status"

test: start-dev gen
	@executor run -d "Running tests" -c "go test ./..."
	@make stop-dev

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
