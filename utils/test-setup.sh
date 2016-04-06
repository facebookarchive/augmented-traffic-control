#!/usr/bin/bash

if [ "$(whoami)" != "root" ] ; then
    sudo "$0" 
    exit 0
fi

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

# Host interface for communication into the ATC netns
ip link add host0 type veth peer name host1

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
ip link set host1 up

# Set out inside interfaces up
ip netns exec cli ip link set cli0 up
ip netns exec atc ip link set lan0 up
ip netns exec atc ip link set wan0 up
ip netns exec srv ip link set srv0 up
ip link set host0 up

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
ip link set dev host1 promisc on

# set master to appropriate bridge
ip link set dev cli1 master br0
ip link set dev lan1 master br0
ip link set dev host1 master br0
ip link set dev wan1 master br1
ip link set dev srv1 master br1

# Enable ip forwarding for both IPv6 and IPv4.
ip netns exec atc sysctl -w net.ipv4.ip_forward=1
ip netns exec atc sysctl -w net.ipv6.conf.all.forwarding=1

# IPv4 Client network: 192.168.3.0/24
# IPv4 Server network: 192.168.4.0/24
# IPv6 Client network: fc00:1::0/32
# IPv6 Server network: fc00:2::0/32

# assign IPv4 addresses to inside interfaces
ip netns exec cli ip -4 addr add dev cli0 192.168.3.2/24 broadcast 192.168.3.255
ip netns exec atc ip -4 addr add dev lan0 192.168.3.1/24 broadcast 192.168.3.255
ip netns exec atc ip -4 addr add dev wan0 192.168.4.1/24 broadcast 192.168.4.255
ip netns exec srv ip -4 addr add dev srv0 192.168.4.2/24 broadcast 192.168.4.255
ip -4 addr add dev host0 192.168.3.3/24 broadcast 192.168.3.255

# assign IPv6 addresses to inside interfaces
ip netns exec cli ip -6 addr add dev cli0 fc00:1::2/32
ip netns exec atc ip -6 addr add dev lan0 fc00:1::1/32
ip netns exec atc ip -6 addr add dev wan0 fc00:2::1/32
ip netns exec srv ip -6 addr add dev srv0 fc00:2::2/32
ip -6 addr add dev host0 fc00:1::3/32

# Add routes so that out-of-network IPs will be forwarded.
ip netns exec cli ip -4 route add 192.168.4.0/24 via 192.168.3.1 dev cli0
ip netns exec srv ip -4 route add 192.168.3.0/24 via 192.168.4.1 dev srv0
ip netns exec cli ip -6 route add fc00:2::0/32 via fc00:1::1 dev cli0
ip netns exec srv ip -6 route add fc00:1::0/32 via fc00:2::1 dev srv0
ip -4 route add 192.168.4.0/24 via 192.168.3.1 dev host0
ip -6 route add fc00:2::0/32 via fc00:1::1 dev host0
