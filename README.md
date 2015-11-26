# Augmented Traffic Control

[![GoDoc](https://godoc.org/github.com/facebook/augmented-traffic-control?status.svg)](https://godoc.org/github.com/facebook/augmented-traffic-control)
[![Build Status](https://travis-ci.org/facebook/augmented-traffic-control.svg?branch=master)](https://travis-ci.org/facebook/augmented-traffic-control)

## WARNING

This version of ATC is experimental and not yet officially released. We welcome
feedback on this version of ATC as we continue working on it, but you should
not use this version of ATC unless you are prepared to deal with bugs or
feature changes.


Full documentation for the project is available at
[http://facebook.github.io/augmented-traffic-control/](http://facebook.github.io/augmented-traffic-control/).


## Overview

Augmented Traffic Control (ATC) is a tool to simulate network conditions. It allows controlling the connection that a device has to the internet.
Developers can use `ATC` to test their application across varying network conditions, easily emulating mobile, and severely impaired networks.

Aspects of the connection that can be controlled include:

* bandwidth
* latency
* packet loss
* corrupted packet rates
* packet ordering

In order to be able to shape the network traffic, ATC must be running on a device that routes the traffic and sees the real IP address of the device,
for example, your network gateway. This also allows any devices that route through `ATC` to be able to shape their traffic.
Traffic can be shaped/unshaped using a web interface allowing any devices with a web browser to use `ATC` without the need for a client application.

ATC is made of two components that interact together:
* `atcd`: A low-level thrift interface which is responsible for directly setting/unsetting traffic shaping.
* `atc_api`: A RESTful HTTP interface to `atcd`. Proxies requests to an `atcd` thrift server.

![ATC architecture][atc_architecture]


## Installing ATC

The fact that `ATC` is split in multiple packages allows for multiple deployment scenarios.
However, deploying all the packages on the same host is the simplest way to setup ATC.

### Requirements

ATC is only supported on **linux** systems. We are considering adding BSD (and OSX) support, but
as of this time, ATC must be run on a linux system.

ATC requires a  working [Golang](https://golang.org/) toolchain.

Most linux distributions have binary packages available for go.

For example:

- *Ubuntu*: `golang`
- *Arch Linux*: `go`

### Installing from Source

First, setup a GOPATH. If you aren't doing normal golang development, you can run
`setup.sh` within the ATC project and a GOPATH will be setup locally specifically for ATC.
After running the script, you must `export GOPATH="$(pwd)/.gopath"` before compiling ATC.

Clone the ATC code somewhere on your machine and run `make` within the project:

``` bash
git clone git@github.com:facebook/augmented-traffic-control.git
cd augmented-traffic-control
./setup.sh
export GOPATH="$(pwd)/.gopath"
make
```

This will download all of ATC's dependencies and compile `atcd` and `atc_api` into `bin/`.

Note that downloading the dependencies may take a while depending on your connection speed.

You can then run `sudo make install` to copy these binaries into `/usr/local/bin/`, but this isn't required to run ATC.

From here, navigate your web browser to `localhost:8080` (replacing `localhost` with the ip address or hostname of your ATC box).


## Building API static resources

```
cd src/react
npm update
npm run build-jsx
npm run build-js
```
