#!/bin/bash

set -e

source /etc/profile.d/node.sh
# execute the command
exec "$@"
