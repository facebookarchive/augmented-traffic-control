#!/bin/bash

export ATCD_ADDR="json://192.168.3.1:9090"

./bin/atc create 192.168.3.2

echo

# 2G developing urban
./bin/atc shape 1 \
  --dn.rate 35 \
  --dn.delay 650 \
  --up.rate 34 \
  --up.delay 650

echo

# execute test
iperf3 -c 192.168.4.2

# dump details

echo

cat log/atcd.log | grep Buffer

echo

ip netns exec atc tc qdisc show dev wan0
ip netns exec atc tc qdisc show dev lan0

echo

ip netns exec atc tc class show dev wan0
ip netns exec atc tc class show dev lan0

echo

ip netns exec atc tc filter show dev wan0
ip netns exec atc tc filter show dev lan0
