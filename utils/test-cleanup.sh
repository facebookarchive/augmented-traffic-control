#!/usr/bin/bash

if [ "$(whoami)" != "root" ] ; then
    sudo "$0"
    exit 0
fi

killall atc_api
killall atcd

# Remove outside interfaces
ip link del dev cli1
ip link del dev lan1
ip link del dev wan1
ip link del dev srv1

# Remove netns
ip netns del atc
ip netns del srv

# Delete bridges
ip link del br0
ip link del br1

# Clean up atc databases
rm -f /tmp/atcd.db /tmp/atc_api.db
