#!/bin/bash

set -e

DIRSH=$(dirname $(realpath "${BASH_SOURCE[0]}"))

PIDFILE="hole-maker.pid"

start-stop-daemon -S -v -b -m -p "${DIRSH}/${PIDFILE}" -x "${DIRSH}/run-hole-maker.sh"
