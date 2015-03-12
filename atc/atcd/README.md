# ATCD

## Introduction

ATCD is the `Augmented Traffic Control` (ATC) Daemon which is reponsible for
handling traffic shaping request for the devices.

`atcd` is written in python and provide a thrift service interface to interact with it.

## Requirements

In order to be able to shape traffic, `atcd` needs to be running on a router
that forwards the packets of your devices.

`atcd` works at Layer 3 so it does shape traffic on a per IP basis, as such, the
`atcd` gateway **must** see the real IP of the devices. In other words, if you are using
NATting, all the devices behind the NAT will get shaped using the same shaping rules.

`atcd` depends on the following packages:

* pyroute2==0.3.3
* pyotp==1.4.1
* sparts==0.7.1
* atc_thrift==0.0.1

## Installation

```
$ cd path/to/atcd
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

`atcd` init files for debian and rhel based distro can be found in the [chef cookbook](../../chef/atc/files/default/init.d/)
