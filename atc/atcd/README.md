# ATCD

## Introduction

ATCD is the `Augmented Traffic Control` (ATC) Daemon which is reponsible for
handling traffic shaping request for the devices.

`atcd` is written in python and provide a [Thrift](https://thrift.apache.org/) interface to interact with it.

## Requirements

In order to be able to shape traffic, `atcd` must be running on a router that forwards the packets of your devices.

`atcd` works at Layer 3 so it does shape traffic on a per IP basis, as such, the
`atcd` gateway **must** see the real IP of the devices. In other words, if you are using
NATting, all the devices behind the NAT will get shaped using the same shaping rules.

`atcd` depends on the following packages:

* python 2.7
* pyroute2==0.3.3
* pyotp==1.4.1
* sparts==0.7.1
* atc_thrift==0.1.3

## Installation

The easiest way to install `atcd` is to install it directly from [pip](https://pypi.python.org/pypi).

### From pip

``` bash
pip install atcd
```

### From source

``` bash
cd path/to/atcd
pip install .
```

## Configuration

`atcd` is configured via command line arguments, to get the full list of options
run:

```
atcd -h
```

The most important options to configure are:

* --atcd-wan: The interface used to connect to internet.
* --atcd-lan: The interface used to connect to your devices.
* --sqlite-file: The location where atcd will keep current device settings.

`atcd` init files for debian and rhel based distro can be found in the [chef cookbook](../../chef/atc/files/default/init.d/)

## How atcd works

### Overview

In order to shape traffic, `atcd` leverages Linux's builtin [Traffic Control subsystem][tchowto]. Communication with the Traffic Control subsystem is done over the netlink API and facilitated by [pyroute2][pyroute2], a pure python netlink library.

Packets that needs to be shaped are expected to be marked. Based on that mark, a classifier will put the packets in the right "buckets", which then will throttle the bandwith, add latency, drop packets, corrupt them... depending on the shaping settings.

The diagram below illustrate the flow an IP packet goes through:
![ATC Diagram][atc_diagram]

### In more details

#### Marking packets

Packets are marked by using iptables'`MARK` target within the `mangle` table. Marking is done as the packet traverses the router on the `FORWARD` chain, e.g when shaping packets for/to IPs 10.0.2.2, 10.0.2.4 and 10.0.2.5:

```bash
-A FORWARD -d 10.0.2.2/32 -i eth0 -j MARK --set-xmark 0x2/0xffffffff
-A FORWARD -s 10.0.2.2/32 -i eth1 -j MARK --set-xmark 0x2/0xffffffff
-A FORWARD -d 10.0.2.4/32 -i eth0 -j MARK --set-xmark 0x3/0xffffffff
-A FORWARD -s 10.0.2.4/32 -i eth1 -j MARK --set-xmark 0x3/0xffffffff
-A FORWARD -d 10.0.2.5/32 -i eth0 -j MARK --set-xmark 0x4/0xffffffff
-A FORWARD -s 10.0.2.5/32 -i eth1 -j MARK --set-xmark 0x4/0xffffffff
```
where `eth0` is the network interface that connects to the internet.

#### Shaping packets

The Traffic Controlling part is more complex. Below is what the shaping on the uplink may look like for 3 devices with IPs 10.0.2.2, 10.0.2.4 and 10.0.2.5:
```bash
# tc filter show dev eth0
filter parent 1: protocol ip pref 1 fw
filter parent 1: protocol ip pref 1 fw handle 0x2 classid 1:2  police 0x1 rate 100000bit burst 12000b mtu 2Kb action drop overhead 0b
ref 1 bind 1

filter parent 1: protocol ip pref 1 fw handle 0x3 classid 1:3  police 0x3 rate 200000bit burst 12000b mtu 2Kb action drop overhead 0b
ref 1 bind 1

filter parent 1: protocol ip pref 1 fw handle 0x4 classid 1:4  police 0x5 rate 200000bit burst 12000b mtu 2Kb action drop overhead 0b
ref 1 bind 1

# tc class show dev eth0
class htb 1:4 root leaf 8005: prio 0 rate 200000bit ceil 200000bit burst 1600b cburst 1600b
class htb 1:2 root leaf 8001: prio 0 rate 100000bit ceil 100000bit burst 1600b cburst 1600b
class htb 1:3 root leaf 8003: prio 0 rate 200000bit ceil 200000bit burst 1600b cburst 1600b
# tc qdisc show dev eth0
qdisc htb 1: root refcnt 2 r2q 10 default 0 direct_packets_stat 3755 direct_qlen 1000
qdisc netem 8001: parent 1:2 limit 1000 delay 10.0ms loss 1%
qdisc netem 8003: parent 1:3 limit 1000 delay 20.0ms loss 2%
qdisc netem 8005: parent 1:4 limit 1000 delay 20.0ms loss 2%
```

So what goes on? When a packets gets in, it goes through the root qdisc (line 4), which is virtually unlimited, filters are checked and if a packet is marked with mark 0x2, it will be passed onto the class with id 1:2 where throttling happens. After that, the packet is passed to its child qdisc that uses netem to provide packet loss, corruption, reordering... and then off it goes on the wire.

The diagram below represents how the `filter`, `class` and `qdisc` fit together:

```
                   root 1:
                    _ |_          <-- filter matching
                   /  |  \
                  /   |   \
                 /    |    \
               1:2   1:3   1:4    <-- bandwidth limits
                |     |     |
              8001: 8003:  8005:  <-- delay, packet loss, reordering and corruption
```

When requesting `atcd` to shape/unshape traffic for a given device, `atcd` will set/unset the needed `iptables` rules and `filter`, `class` and `qdisc` to control the traffic. Aside from this, it will run some periodic tasks for housekeeping (like expiring shaping settings...).

## Security

`atcd` has currently almost no authentication/authorization mechanism built-in. It is recommended to make `atcd` only listen on `localhost`, and offload the authentication to the API.

[tchowto]: http://www.tldp.org/HOWTO/Traffic-Control-HOWTO/
[pyroute2]: https://github.com/svinota/pyroute2
[atc_diagram]: https://facebook.github.io/augmented-traffic-control/images/atc_diagram.png
