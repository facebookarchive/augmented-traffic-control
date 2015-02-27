#!/usr/bin/env python
#
#  Copyright (c) 2014, Facebook, Inc.
#  All rights reserved.
#
#  This source code is licensed under the BSD-style license found in the
#  LICENSE file in the root directory of this source tree. An additional grant
#  of patent rights can be found in the PATENTS file in the same directory.
#
#
from atc_thrift import Atcd
from atc_thrift.ttypes import TrafficControl
from thrift.protocol import TBinaryProtocol
from thrift.transport import TSocket, TTransport

from atc_thrift.ttypes import Shaping
from atc_thrift.ttypes import TrafficControlledDevice
from atc_thrift.ttypes import TrafficControlSetting

import argparse


def getAtcClient():
        transport = TSocket.TSocket('localhost', 9090)
        transport = TTransport.TFramedTransport(transport)
        transport.open()
        protocol = TBinaryProtocol.TBinaryProtocol(transport)
        return Atcd.Client(protocol)


def parse_arguments():
    parser = argparse.ArgumentParser()
    parser.add_argument(
        '--self',
        action='store_true',
        help='Shape for oneself?'
    )
    parser.add_argument(
        '--controlling-ip',
        default='1.1.1.1',
        help='Controlling ip [%(default)s]'
    )
    parser.add_argument(
        '--controlled-ip',
        default='2.2.2.2',
        help='Controlled ip [%(default)s]'
    )
    return parser.parse_args()

if __name__ == '__main__':
    options = parse_arguments()
    client = getAtcClient()
    dev = TrafficControlledDevice(
        controllingIP=options.controlling_ip,
        controlledIP=options.controlling_ip if options.self
        else options.controlled_ip
    )
    settings = TrafficControlSetting(
        up=Shaping(
            rate=100,
        ),
        down=Shaping(
            rate=200,
        ),
    )
    print settings
    tc = TrafficControl(
        device=dev,
        settings=settings,
        timeout=1000,
    )

    print client.startShaping(tc)
