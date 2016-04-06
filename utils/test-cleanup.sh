#!/usr/bin/bash

if [ "$(whoami)" != "root" ] ; then
    sudo "$0"
    exit 0
fi

# Remove outside interfaces
ip link del dev cli1
ip link del dev lan1
ip link del dev wan1
ip link del dev srv1
ip link del dev host0
ip link del dev host1

# Remove netns
ip netns del cli
ip netns del atc
ip netns del srv

# Delete bridges
ip link del br0
ip link del br1
ip link del br1
