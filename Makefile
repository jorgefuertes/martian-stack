SHELL=/usr/bin/env bash

gen:
	@go tool executor run -d "Generating templates" -c "go tool goht generate"

start-dev:
	@go tool executor run -d "Starting Redis" -c "scripts/pod.sh redis start"

stop-dev:
	@go tool executor run -d "Stopping Redis" -c "scripts/pod.sh redis stop"

status-dev:
	@go tool executor run -d "Checking Redis" -c "scripts/pod.sh redis status"

test: start-dev gen
	@(set -e; err=0; \
		go tool executor run -d "Running tests" -c "go test ./..." || err=$$?; \
		make stop-dev; \
		exit $$err)

test-clean:
	@go tool executor run -d "Cleaning test cache" -c "go clean -testcache"
	@make test

lint:
	@go tool executor run -d "staticcheck" -c "go tool staticcheck ./..."
	@go tool executor run -d "gofumpt" -c "go tool gofumpt -d -l -extra ."
	@go tool executor run -d "golines" -c "go tool golines -w -m 120 --no-reformat-tags ."
	@go tool executor run -d "vet" -c "go vet ./..."
	@go tool executor run -d "golangci-lint" -c "GOGC=80 go tool golangci-lint run --fast --concurrency 16"
	@go tool executor run -d "govulncheck" -c "go tool govulncheck ./..."
	@go tool executor run -d "markdownlint" -c "trunk check --filter=markdownlint --all"

run: start-dev gen
	go run cmd/testserver/main.go
	@make stop-dev
