#
#  Copyright (c) 2014, Facebook, Inc.
#  All rights reserved.
#
#  This source code is licensed under the BSD-style license found in the
#  LICENSE file in the root directory of this source tree. An additional grant
#  of patent rights can be found in the PATENTS file in the same directory.
#
#

BITS = long(1)
KILOBITS = BITS * 1024
MEGABITS = KILOBITS * 1024
GIGABITS = MEGABITS * 1024


class Speed(object):
    value = 0

    def __init__(self, value, unit=BITS):
        self.value = long(value * unit)

    def faster(self, other):
        return self.value > other.value

    def slower(self, other):
        return self.value < other.value

    def kbps(self):
        return int(self.value / KILOBITS)

    def __str__(self):
        i = 0
        value = self.value
        while i < 3 and value % KILOBITS == 0:
            i += 1
            value = value / 1024
        if i == 0:
            return str(value) + ' bits/sec'
        if i == 1:
            return str(value) + ' Kbits/sec'
        if i == 2:
            return str(value) + ' Mbits/sec'
        return str(value) + ' Gbits/sec'

    def __add__(self, other):
        if isinstance(other, Speed):
            return Speed(self.value + other.value)
        else:
            return Speed(self.value + other)

    def __sub__(self, other):
        if isinstance(other, Speed):
            return Speed(self.value - other.value)
        else:
            return Speed(self.value - other)

    def __div__(self, other):
        if isinstance(other, Speed):
            return Speed(self.value / other.value)
        else:
            return Speed(self.value / other)

    def __mul__(self, other):
        if isinstance(other, Speed):
            return Speed(self.value * other.value)
        else:
            return Speed(self.value * other)


Kilobit = Speed(1, KILOBITS)
Megabit = Speed(1, MEGABITS)
Gigabit = Speed(1, GIGABITS)


def parseIPerfSpeed(s):
    speeds, unit = s.split()[-2:]
    try:
        speed = float(speeds)
    except:
        print 'Invalid speed line: ' + repr(s)
        raise
    if speed > GIGABITS:
        # something is likely fishy with this value.
        # Print it incase the test fails.
        print 'bad iperf speed(?):', str(s)
    if unit == 'bits/sec':
        return Speed(speed)
    if unit == 'Bytes/sec':
        return Speed(speed, 8)
    if unit == 'Kbits/sec':
        return Speed(speed, KILOBITS)
    if unit == 'Mbits/sec':
        return Speed(speed, MEGABITS)
    if unit == 'Gbits/sec':
        return Speed(speed, GIGABITS)
    print 'Invalid speed line: ' + repr(s)
    raise ValueError('Unknown unit for network speed: ' + unit)
