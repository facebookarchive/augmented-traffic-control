#!/bin/sh
#
# atcd - this script starts and stops the atcd traffic shaping daemon
#
# config: /etc/default/atcd /etc/atcd.conf
# pidfile: /var/run/atcd.pid
#
### BEGIN INIT INFO
# Provides:       atcd
# Required-Start:
# Required-Stop:
# Default-Start:  2 3 4 5
# Default-Stop:   0 1 6
# Description:    ATCD is a thrift service that applies traffic shaping rules
### END INIT INFO

# Source function library
. /lib/lsb/init-functions

PATH="/usr/local/sbin:/usr/local/bin:${PATH}"
name="atcd"

ATCD_LISTEN_ADDRESS=127.0.0.1
ATCD_LISTEN_PORT=9090
ATCD_VENV=
ATCD_SQLITE=/var/lib/atc/atcd.db
ATCD_WAN=eth0
ATCD_LAN=eth1
ATCD_MODE=secure

sysconfig="/etc/default/${name}"
lockfile="/var/lock/${name}"
pidfile="/var/run/${name}.pid"

[ -f $sysconfig ] && . $sysconfig

start() {
    [ ${ATCD_VENV} ] && . ${ATCD_VENV}/bin/activate
    atcd="$(which ${name})"
    [ -x $atcd ] || exit 5
    echo -n "Starting $name: "
    start-stop-daemon --start --pidfile $pidfile --exec $atcd -- --pidfile $pidfile --daemon --thrift-host ${ATCD_LISTEN_ADDRESS} --thrift-port ${ATCD_LISTEN_PORT} --sqlite-file ${ATCD_SQLITE} --atcd-lan ${ATCD_LAN} --atcd-wan ${ATCD_WAN} --atcd-mode ${ATCD_MODE}
    retval=$?
    echo
    [ $retval -eq 0 ] && touch $lockfile
    return $retval
}

stop() {
    echo -n "Stopping $name: "
    start-stop-daemon --stop --pidfile $pidfile
    retval=$?
    echo
    [ $retval -eq 0 ] && rm -f $lockfile
    return $retval
}

restart() {
    stop
    start
}

case "$1" in
    start)
        $1
        ;;
    stop)
        $1
        ;;
    restart)
        $1
        ;;
    condrestart|try-restart)
        rh_status_q || exit 7
        restart
	    ;;
    *)
        echo $"Usage: $0 {start|stop|restart}"
        exit 2
esac
