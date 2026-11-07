#!/bin/bash
ENDPOINT="http://localhost:12120"
CT="Content-Type: application/json"
AC="Accept: application/json"

if [ $# -ne 1 ]; then
	echo "test_apis.sh type"
	exit 1
fi

if [ "$1" == "lock_office_green" ]; then

	curl -v ${ENDPOINT}/lock -H "${CT}" -H "${AC}" \
	-d '{"targets":["office_lights"],"for":"5s","color":"#00ff00"}'
	exit 1
fi

if [ "$1" == "office_warm" ]; then

	curl -v ${ENDPOINT}/addscenes -H "${CT}" -H "${AC}" \
	-d '{"scenes":["office_warm"],"clear_existing":true,"for":"1m"}'
	exit 1
fi
	

echo "provide a valid type"
