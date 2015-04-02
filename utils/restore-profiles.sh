#!/usr/bin/env bash

ATC_HOST="$1"

if [ -z "$ATC_HOST" ] ; then
    echo "usage: $0 HOST"
    echo "    HOST should be of the form <hostname>:<port>"
    echo "    e.g. localhost:8080"
    exit 1
fi

for x in $(git rev-parse --show-toplevel)/utils/profiles/* ; do
    echo "Adding profile $x"
    curl --silent http://"$ATC_HOST"/api/v1/profiles/ -d "@${x}" 1>/dev/null
done
