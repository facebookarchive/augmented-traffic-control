#!/bin/bash

set -e

# start up rsyslog
/etc/init.d/rsyslog start
# execute the command
exec "$@"
