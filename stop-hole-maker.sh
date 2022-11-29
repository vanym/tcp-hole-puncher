#!/bin/bash

set -e

HOLE_MAKER_PID_FILE="${HOLE_MAKER_PID_FILE:-hole-maker.pid}"

start-stop-daemon -K -v -p "${HOLE_MAKER_PID_FILE}"
