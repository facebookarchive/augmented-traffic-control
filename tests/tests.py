#
#  Copyright (c) 2014, Facebook, Inc.
#  All rights reserved.
#
#  This source code is licensed under the BSD-style license found in the
#  LICENSE file in the root directory of this source tree. An additional grant
#  of patent rights can be found in the PATENTS file in the same directory.
#
#

import unittest

from vagrant import Vagrant
from vms import speedBetween, shape, unshape
from speed import Megabit


IPERF_OPTS = {
    'udp': False,
    'time': 30
}


class TestAtcdE2E(unittest.TestCase):

    def test_shapesBandwidth(self):
        '''
        Tests that bandwidth shaping works.

        Examines the network speed before, during, and after shaping.

        Fails if the network speeds do not reflect expected results.
        '''
        with Vagrant.ssh('gateway', 'client', 'server') as machines:
            gateway, client, server = machines

            before = speedBetween(client, server, **IPERF_OPTS)

            print 'Actual speed before shaping:', before

            shapedSpeed = before / 1024

            print 'Desired shaping speed:', shapedSpeed

            shape(gateway, client, shapedSpeed)

            during = speedBetween(client, server, **IPERF_OPTS)

            print 'Actual speed during shaping:', during

            unshape(gateway, client)

            after = speedBetween(client, server, **IPERF_OPTS)

            print 'Actual speed after shaping:', after

            if before.slower(Megabit):
                self.fail(
                    'Actual speed before shaping is too slow for shape'
                    ' testing. (is it already being shaped?)')

            if during.faster(shapedSpeed):
                self.fail(
                    'Actual speed during shaping exceeded'
                    ' shaping speed.')

            if after.slower(shapedSpeed * 2):
                self.fail(
                    'Actual speed after shaping appears'
                    ' to still be shaped.')

            if after.slower(before * 0.7):
                self.fail(
                    'Actual speed after shaping did not return'
                    ' to normal value after shaping.')
