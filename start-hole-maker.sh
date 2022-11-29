#!/bin/bash

set -e

DIRSH=$(dirname $(realpath "${BASH_SOURCE[0]}"))

HOLE_MAKER_PID_FILE="${HOLE_MAKER_PID_FILE:-hole-maker.pid}"

start-stop-daemon -S -v -b -d "$PWD" -m -p "${HOLE_MAKER_PID_FILE}" -x "${DIRSH}/run-hole-maker.sh"
