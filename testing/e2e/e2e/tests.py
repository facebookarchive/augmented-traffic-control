import unittest
import sys

from e2e.vagrant import Vagrant
from e2e.vms import speedBetween, shape, unshape
from e2e.speed import Kilobit, Megabit


class TestAtcdE2E(unittest.TestCase):
    ShapedSpeed = Kilobit * 52

    # If two network speeds are within this margin (%)
    # from one another they are considered the same.
    Margin = 0.15

    def test_shapesBandwidth(self):
        '''
        Tests that bandwidth shaping works.

        Examines the network speed before, during, and after shaping.
        Fails if:
            - Speed before or after shaping is being shaped.
            - Speed during shaping is not being shaped.
        '''
        with Vagrant.ssh('gateway', 'client', 'server') as machines:
            gateway, client, server = machines

            before = speedBetween(client, server)

            if self.ShapedSpeed.withinMargin(self.Margin, before):
                self.fail(
                    "Actual speed (" + str(before) +
                    ") is too slow for shape testing. " +
                    "(is it already being shaped?)")

            shape(gateway, client, self.ShapedSpeed)

            during = speedBetween(client, server)

            if not self.ShapedSpeed.withinMargin(self.Margin, during):
                self.fail(
                    "Actual speed (" + str(during) +
                    ") did not change during shaping.")

            unshape(gateway, client)

            after = speedBetween(client, server)

            if self.ShapedSpeed.withinMargin(self.Margin, after):
                self.fail(
                    "Actual speed (" + str(after) +
                    ") did not change after shaping.")

            if not before.withinMargin(self.Margin, after):
                self.fail(
                    "Actual speed (" + str(after) +
                    ") did not change after shaping.")

    def test_dropsPackets(self):
        pass
