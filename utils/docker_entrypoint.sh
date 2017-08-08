#!/bin/bash

set -e

# start up rsyslog
/etc/init.d/rsyslog start
# set up testing network namespaces
bash utils/test-setup.sh
# execute the command
exec "$@"
