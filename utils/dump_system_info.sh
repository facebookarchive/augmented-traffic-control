#!/bin/bash

WAN=${1:-eth0}
LAN=${2:-eth1}

function title {
    echo
    echo "######### $1 #########"
    echo
}

function run_cmd {
    title "$1"
    $@
}

function run_cmd_filtered {
    title "$1"
    $@ | egrep -v '(ether|inet6.*scope link)'
}

for cmd in "uname -a" "cat /etc/os-release" "python -V" "pip freeze" \
    "ip r" "iptables-save" "ip r"
do
    run_cmd "${cmd}"
done

run_cmd_filtered "ip a"

for netif in ${WAN} ${LAN}
do
    for object in qdisc class filter
    do
        cmd="tc ${object} show dev ${netif}"
        run_cmd "${cmd}"
    done
done
