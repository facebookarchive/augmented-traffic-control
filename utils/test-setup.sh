#!/bin/bash

if [ "$(whoami)" != "root" ] ; then
    sudo "$0"
    exit 0
fi

ATC_ROOT="$(dirname $(dirname $(realpath $0)))"
ATCD_PATH="$ATC_ROOT/bin/atcd"
ATC_API_PATH="$ATC_ROOT/bin/atc_api"

# Create the network namespaces
ip netns add atc
ip netns add srv

# Bring up loopback interfaces inside the network namespaces
ip netns exec atc ip link set lo up
ip netns exec srv ip link set lo up

# Create paired veth interfaces
# x0 is inside, x1 is outside
ip link add lan0 type veth peer name lan1
ip link add wan0 type veth peer name wan1
ip link add cli0 type veth peer name cli1
ip link add srv0 type veth peer name srv1

# Assign the inside interfaces to the network namespaces
# cli0 is on the host
ip link set dev lan0 netns atc
ip link set dev wan0 netns atc
ip link set dev srv0 netns srv

# Set our outside interfaces up
ip link set cli1 up
ip link set lan1 up
ip link set wan1 up
ip link set srv1 up

# Set out inside interfaces up
ip link set cli0 up
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

# Enable ip forwarding for both IPv6 and IPv4.
ip netns exec atc sysctl -w net.ipv4.ip_forward=1 >/dev/null
ip netns exec atc sysctl -w net.ipv6.conf.all.forwarding=1 >/dev/null

# IPv4 Client network: 192.168.3.0/24
# IPv4 Server network: 192.168.4.0/24
# IPv6 Client network: fc00:1::0/32
# IPv6 Server network: fc00:2::0/32

# assign IPv4 addresses to inside interfaces
ip -4 addr add dev cli0 192.168.3.2/24 broadcast 192.168.3.255
ip netns exec atc ip -4 addr add dev lan0 192.168.3.1/24 broadcast 192.168.3.255
ip netns exec atc ip -4 addr add dev wan0 192.168.4.1/24 broadcast 192.168.4.255
ip netns exec srv ip -4 addr add dev srv0 192.168.4.2/24 broadcast 192.168.4.255

# assign IPv6 addresses to inside interfaces
if [ -n "$ENABLE_IPV6" ]; then
	ip -6 addr add dev cli0 fc00:1::2/32
	ip netns exec atc ip -6 addr add dev lan0 fc00:1::1/32
	ip netns exec atc ip -6 addr add dev wan0 fc00:2::1/32
	ip netns exec srv ip -6 addr add dev srv0 fc00:2::2/32
fi

# Add routes so that out-of-network IPs will be forwarded.
ip -4 route add 192.168.4.0/24 via 192.168.3.1 dev cli0
ip netns exec srv ip -4 route add 192.168.3.0/24 via 192.168.4.1 dev srv0

if [ -n "$ENABLE_IPV6" ]; then
	ip -6 route add fc00:2::0/32 via fc00:1::1 dev cli0
	ip netns exec srv ip -6 route add fc00:1::0/32 via fc00:2::1 dev srv0
fi

if [ -n "$RESET_DB" ]; then
  echo "Resetting database"
  rm db/*.db
fi

if [ -n "$RESET_LOGS" ]; then
  echo "Resetting logs"
  rm log/*.log
fi

mkdir -p $ATC_ROOT/log
mkdir -p $ATC_ROOT/db

ip netns exec srv /usr/bin/iperf3 -d -V -s &>$ATC_ROOT/log/iperf3.log &
ip netns exec atc $ATCD_PATH --wan wan0 --lan lan0 -Q "$ATC_ROOT/db/atcd.db" --insecure -v -b 0.0.0.0:9090 &>$ATC_ROOT/log/atcd.log &
sleep 3 # wait for atcd to startup
$ATC_API_PATH -t json://192.168.3.1:9090 -W -v -4 127.0.0.1 -6 fc00::1::1 -Q "$ATC_ROOT/db/atc_api.db" --assets "$ATC_ROOT/static" &>$ATC_ROOT/log/atc_api.log &

GRN="\e[32m"
CLR="\e[39m"

echo -e "${GRN}Setup Successful!${CLR}"

echo -e "${GRN}In order to use the ATC cli use either of:${CLR}"
echo "export ATCD_ADDR=\"json://192.168.3.1:9090\""
echo "export ATCD_ADDR=\"json://[fc00:1::1]:9090\""

echo -e "${GRN}ATC Web UI:${CLR}"
echo "http://192.168.3.1:8080"
echo "http://[fc00:1::1]:8080"

echo -e "${GRN}To test with iperf3:${CLR}"
echo "iperf3 -c 192.168.4.2"
echo "iperf3 -c fc00:2::2"
