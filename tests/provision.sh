#!/bin/bash -eu

if [[ "$(hostname)" == "gateway" ]] ; then
    sed -i -e 's/ATCD_WAN=eth[012]/ATCD_WAN=eth1/' -e 's/ATCD_LAN=eth[012]/ATCD_LAN=eth2/' /etc/default/atcd

    /etc/init.d/atcd restart

    /usr/local/bin/atcui-setup
else
    if [[ "$(hostname)" == "client" ]] ; then
        ip route add 192.168.10.0/24 via 192.168.20.2
    elif [[ "$(hostname)" == "server" ]] ; then
        ip route add 192.168.20.0/24 via 192.168.10.2
    fi

    apt-get install iperf
fi
