#!/bin/bash

set -e

DIRSH=$(dirname $(realpath "${BASH_SOURCE[0]}"))

PIDFILE="hole-maker.pid"

start-stop-daemon -K -v -p "${DIRSH}/${PIDFILE}"
