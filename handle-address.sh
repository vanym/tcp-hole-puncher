#!/bin/bash

set -e

DIRSH=$(dirname $(realpath "${BASH_SOURCE[0]}"))

FULL_ADDRESS="$1"

IP_PORT_EXP='(.+):([0-9]+)'
[[ "${FULL_ADDRESS}" =~ $IP_PORT_EXP ]]

IP="${BASH_REMATCH[1]}"
PORT="${BASH_REMATCH[2]}"

echo "${IP}:${PORT}" | tee "${DIRSH}"/address.txt
