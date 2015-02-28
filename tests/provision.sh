#!/bin/bash -eu

#
#  Copyright (c) 2014, Facebook, Inc.
#  All rights reserved.
#
#  This source code is licensed under the BSD-style license found in the
#  LICENSE file in the root directory of this source tree. An additional grant
#  of patent rights can be found in the PATENTS file in the same directory.
#
#

if [[ "$(hostname)" == "gateway" ]] ; then
    sed -i -e 's/ATCD_WAN=eth[012]/ATCD_WAN=eth1/' -e 's/ATCD_LAN=eth[012]/ATCD_LAN=eth2/' /etc/default/atcd

    /etc/init.d/atcd restart

    /usr/local/bin/atcui-setup
else
    if [[ "$(hostname)" == "client" ]] ; then
        ip route add 192.168.10.0/24 via 192.168.20.2
    elif [[ "$(hostname)" == "server" ]] ; then
        ip route add 192.168.20.0/24 via 192.168.10.2
    fi

    apt-get install iperf
fi
