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
	pushd martian-data && go test ./... && popd
	pushd martian-http && go test ./... && popd
