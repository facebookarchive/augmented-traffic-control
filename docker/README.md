# docker-atc

A Dockerfile that installs the atc stack.

# Installation

The easiest way to get this image is to run
```
docker run -it --privileged --net=host --rm atcd/atc
```

# Building the image

If you prefer to build the image yourself:
```bash
git clone https://github.com/facebook/augmented-traffic-control.git
cd augmented-traffic-control/docker
docker build -t="${USER}/atc" .
```

# Usage

`ATC` modifies the network of the host it is running on. As such, it **MUST**
run with the following options: `--cap-add=NET_ADMIN --net=host`.

There is currently 2 environment variables:
* ATCD_WAN (default *eth0*)
* ATCD_LAN (default *eth1*)
* ATCD_MODE (default *secure*)
* ATCD_BURST_SIZE (default *12000*)
* ATCD_OPTIONS (default *empty*) free form options to pass to atcd

To run atcd with the default settings:

```
docker run -it --cap-add=NET_ADMIN --net=host --rm atcd/atc
```

To change which interface to use for WAN access (internet) or LAN, you can modify
*ATCD_WAN* and *ATCD_LAN* environment variables. For example:


```
docker run -it --cap-add=NET_ADMIN --net=host --rm -e ATCD_LAN=wlan1 -e ATCD_WAN=em1 atcd/atc
```

You can then access ATC UI at:

```
http://docker_ip:8000
```
