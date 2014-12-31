#
#  Copyright (c) 2014, Facebook, Inc.
#  All rights reserved.
#
#  This source code is licensed under the BSD-style license found in the
#  LICENSE file in the root directory of this source tree. An additional grant
#  of patent rights can be found in the PATENTS file in the same directory.
#
#
from atc_api.settings import atc_api_settings
from atc_thrift import Atcd
from thrift.transport import TSocket, TTransport
from thrift.protocol import TBinaryProtocol


def atcdClient():
    try:
        transport = TSocket.TSocket(
            atc_api_settings.ATCD_HOST,
            atc_api_settings.ATCD_PORT
        )
        transport = TTransport.TFramedTransport(transport)
        transport.open()
        protocol = TBinaryProtocol.TBinaryProtocol(transport)
        return Atcd.Client(protocol)
    except TTransport.TTransportException as e:
        print 'atcdClient: %s: %s' % (e.__class__.__name__, str(e))
