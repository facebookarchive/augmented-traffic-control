#
#  Copyright (c) 2014, Facebook, Inc.
#  All rights reserved.
#
#  This source code is licensed under the BSD-style license found in the
#  LICENSE file in the root directory of this source tree. An additional grant
#  of patent rights can be found in the PATENTS file in the same directory.
#
#
'''
ATC Daemon main file
'''
from __future__ import absolute_import
from __future__ import print_function

import sys

# AtcdHandler main class
from atcd.AtcdDeviceTimeoutTask import AtcdDeviceTimeoutTask
from atcd.AtcdThriftHandlerTask import AtcdNBServerTask
from atcd.AtcdThriftHandlerTask import AtcdThriftHandlerTask
from atcd.AtcdVService import AtcdVService


def initialize_thrift():
    AtcdNBServerTask.register()
    AtcdThriftHandlerTask.factory().register()
    AtcdDeviceTimeoutTask.register()

    AtcdVService.initFromCLI()


def run():
    initialize_thrift()
    sys.exit(0)
