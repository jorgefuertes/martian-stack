SHELL=/usr/bin/env bash

start-dev:
	scripts/pod.sh mongo start
	scripts/pod.sh redis start

stop-dev:
	scripts/pod.sh mongo stop
	scripts/pod.sh redis stop

status-dev:
	scripts/pod.sh mongo status
	scripts/pod.sh redis status

test: start-dev
	go test ./...
	make stop-dev

test-clean:
	go clean -testcache
	make test
