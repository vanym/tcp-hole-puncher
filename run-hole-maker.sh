#!/bin/bash

TCP_HOLE_PORT="${TCP_HOLE_PORT:-7203}"
TCP_HOLE_REDIR_PORT="${TCP_HOLE_REDIR_PORT:-8080}"

set -m

DIRSH=$(dirname $(realpath "${BASH_SOURCE[0]}"))

sudo -n iptables -t nat -A PREROUTING -p tcp --dport "${TCP_HOLE_PORT}" -j REDIRECT --to-ports "${TCP_HOLE_REDIR_PORT}"
if [ $? -ne 0 ]; then
  socat tcp-listen:"${TCP_HOLE_PORT}",reuseaddr,reuseport,fork tcp-connect:0.0.0.0:"${TCP_HOLE_REDIR_PORT}" &
fi

set -e

on_exit(){
  sudo -n iptables -t nat -D PREROUTING -p tcp --dport "${TCP_HOLE_PORT}" -j REDIRECT --to-ports "${TCP_HOLE_REDIR_PORT}"
}
trap on_exit EXIT

on_term(){
  kill -SIGINT $(ps -o pid= --ppid "$$")
}
trap on_term TERM INT

TCP_HOLE_CMD="${TCP_HOLE_CMD:-tcp-hole-puncher}"
if ! which "${TCP_HOLE_CMD}" >&- 2>&- ; then
  TCP_HOLE_CMD="${DIRSH}"/"${TCP_HOLE_CMD}"
fi

"${TCP_HOLE_CMD}" --bind :"${TCP_HOLE_PORT}" "${@}" | stdbuf -i0 -o0 -e0 uniq | xargs -n1 "${TCP_HOLE_ADDRESS_HANDLER:-"${DIRSH}"/handle-address.sh}" &
wait -f
