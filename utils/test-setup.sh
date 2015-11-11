#!/usr/bin/bash

if [ "$(whoami)" != "root" ] ; then
    sudo "$0" 
    exit 0
fi

echo 1 > /proc/sys/net/ipv4/ip_forward

# Create the network namespaces
ip netns add cli
ip netns add atc
ip netns add srv

# Bring up loopback interfaces inside the network namespaces
ip netns exec cli ip link set lo up
ip netns exec atc ip link set lo up
ip netns exec srv ip link set lo up

# Create paired veth interfaces
# x0 is inside, x1 is outside
ip link add lan0 type veth peer name lan1
ip link add wan0 type veth peer name wan1
ip link add cli0 type veth peer name cli1
ip link add srv0 type veth peer name srv1

# Assign the inside interfaces to the network namespaces
ip link set dev cli0 netns cli
ip link set dev lan0 netns atc
ip link set dev wan0 netns atc
ip link set dev srv0 netns srv

# Set our outside interfaces up
ip link set cli1 up
ip link set lan1 up
ip link set wan1 up
ip link set srv1 up

# Set out inside interfaces up
ip netns exec cli ip link set cli0 up
ip netns exec atc ip link set lan0 up
ip netns exec atc ip link set wan0 up
ip netns exec srv ip link set srv0 up

# br0 bridges lan1 and cli1
ip link add br0 type bridge
ip link set dev br0 up

# br1 bridges wan1 and srv1
ip link add br1 type bridge
ip link set dev br1 up

# promiscuous mode required for bridges
ip link set dev cli1 promisc on
ip link set dev lan1 promisc on
ip link set dev wan1 promisc on
ip link set dev srv1 promisc on

# set master to appropriate bridge
ip link set dev cli1 master br0
ip link set dev lan1 master br0
ip link set dev wan1 master br1
ip link set dev srv1 master br1

# assign IP addresses to inside interfaces
ip netns exec cli ip addr add dev cli0 192.168.3.2/24 broadcast 192.168.3.255
ip netns exec atc ip addr add dev lan0 192.168.3.1/24 broadcast 192.168.3.255
ip netns exec atc ip addr add dev wan0 192.168.4.1/24 broadcast 192.168.4.255
ip netns exec srv ip addr add dev srv0 192.168.4.2/24 broadcast 192.168.4.255

# Add default routes so that out-of-network IPs will be forwarded.
ip netns exec cli ip route add default via 192.168.3.1 dev cli0
ip netns exec srv ip route add default via 192.168.4.1 dev srv0
