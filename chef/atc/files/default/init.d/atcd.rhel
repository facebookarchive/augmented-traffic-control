#!/bin/sh
#
# atcd - this script starts and stops the atcd traffic shaping daemon
#
# chkconfig:   2345 80 20
# description:  ATCD is a thrift service that applies traffic shaping rules
# processname: atcd
# config:      /etc/sysconfig/atcd /etc/atcd.conf
# pidfile:     /var/run/atcd.pid
#

# Source function library.
. /etc/rc.d/init.d/functions

# Source networking configuration.
. /etc/sysconfig/network

# Check that networking is up.
[ "$NETWORKING" = "no" ] && exit 0

PATH="/usr/local/sbin:/usr/local/bin:${PATH}"
name="atcd"
prog=${name}

ATCD_LISTEN_ADDRESS=127.0.0.1
ATCD_LISTEN_PORT=9090
ATCD_VENV=
ATCD_SQLITE=/var/lib/atc/atcd.db
ATCD_WAN=eth0
ATCD_LAN=eth1
ATCD_MODE=secure

sysconfig="/etc/sysconfig/${name}"
lockfile="/var/lock/subsys/${name}"
pidfile="/var/run/${name}.pid"

[ -f $sysconfig ] && . $sysconfig


start() {
    [ ${ATCD_VENV} ] && . ${ATCD_VENV}/bin/activate
    atcd="$(which ${name})"
    [ -x $atcd ] || exit 5
    echo -n "Starting $name: "
    daemon $atcd --pidfile $pidfile --daemon --thrift-host ${ATCD_LISTEN_ADDRESS} --thrift-port ${ATCD_LISTEN_PORT} --sqlite-file ${ATCD_SQLITE} --atcd-wan ${ATCD_WAN} --atcd-lan ${ATCD_LAN} --atcd-mode ${ATCD_MODE}
    retval=$?
    echo
    [ $retval -eq 0 ] && touch $lockfile
    return $retval
}

stop() {
    echo -n "Stopping $name: "
    killproc -p $pidfile $prog
    retval=$?
    echo
    [ $retval -eq 0 ] && rm -f $lockfile
    return $retval
}

restart() {
    stop
    start
}

rh_status() {
    status $prog
}

rh_status_q() {
    rh_status >/dev/null 2>&1
}

case "$1" in
    start)
        rh_status_q && exit 0
        $1
        ;;
    stop)
        rh_status_q || exit 0
        $1
        ;;
    restart)
        $1
        ;;
    status|status_q)
        rh_$1
        ;;
    condrestart|try-restart)
        rh_status_q || exit 7
        restart
	    ;;
    *)
        echo $"Usage: $0 {start|stop|status|restart}"
        exit 2
esac
