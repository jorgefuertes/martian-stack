#!/usr/bin/env bash

function helpAndExit {
	echo "Usage: $0 <redis> <start|stop|status>"
	exit 1
}

function okOrFail {
	if [[ $1 -eq 0 ]]; then echo "OK"; else echo "FAIL"; fi
	return "$1"
}

function checkContainer {
	if [[ $2 == "verbose" ]]; then echo -n "Checking ${1}..."; fi
	podman container exists "$1"
	LEVEL=$?
	if [[ $2 == "verbose" ]]; then okOrFail "${LEVEL}"; fi
	return "${LEVEL}"
}

function stopContainer {
	if checkContainer "$1"; then
		echo -n "Stopping ${1}..."
		podman stop "$1" &>/dev/null
		okOrFail $?
	else
		echo "Container $1 is not running"
	fi
	return $?
}

function startContainer {
	checkContainer "$1"
	if [[ $? -ne 0 ]]; then
		DATA_VOLUME="volumes/$1/data"
		mkdir -p "${DATA_VOLUME}"
		if [[ $1 =~ redis$ ]]; then
			# redis
			podman run --rm --name "$1" -v "${PWD}"/"${DATA_VOLUME}":/data -p 6379:6379 -d redis
		else
			echo "Unknown container name"
			helpAndExit
		fi
	else
		echo "Container $1 is already running"
	fi
	return $?
}

if [[ $1 == "help" ]]; then helpAndExit; fi
if [[ $1 == "" ]]; then
	echo "Missing container name"
	helpAndExit
fi
if ! [[ $1 =~ ^redis$ ]]; then
	echo "Unknown container name"
	helpAndExit
fi
if [[ $2 == "" ]]; then
	echo "Missing command"
	helpAndExit
fi
if ! [[ $2 =~ ^start|stop|status$ ]]; then
	echo "Unknown command"
	helpAndExit
fi
if [[ "$(basename "${PWD}")" != "martian-stack" ]]; then
	echo "ERROR: .gitignore not found!"
	echo "Please run this command from the root of martian-stack"
	helpAndExit
fi

PROJECT_NAME=$(basename "${PWD}")
if [[ ${PROJECT_NAME} == "" ]]; then
	echo "ERROR: Cannot get project name"
	helpAndExit
fi

POD_NAME="${PROJECT_NAME}-${1}"

if [[ $2 == "status" ]]; then
	checkContainer "${POD_NAME}" verbose
elif [[ $2 == "stop" ]]; then
	stopContainer "${POD_NAME}"
else
	startContainer "${POD_NAME}"
fi
