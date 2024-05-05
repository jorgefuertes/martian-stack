#!/usr/bin/env bash

function usage {
		echo "Usage: $0 <dir>"
		exit 1
}

function start_dev {
	make start-dev
	if [[ $? -ne 0 ]]; then exit 1; fi
}

function stop_dev {
	make stop-dev
}

if [[ $1 == "" ]]
then
	usage
fi

# dev environment
make status-dev &> /dev/null
if [[ $? -ne 0 ]]
then
	DEV_WAS="STOPPED"
	start_dev
else
	echo "Dev environment already started"
	DEV_WAS="STARTED"
fi

pushd $1 && go test ./...
LEVEL=$?
popd

if [[ "$DEV_WAS" == "STOPPED" ]]
then
	stop_dev
fi

exit $LEVEL
