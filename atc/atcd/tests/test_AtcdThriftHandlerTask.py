#
#  Copyright (c) 2014, Facebook, Inc.
#  All rights reserved.
#
#  This source code is licensed under the BSD-style license found in the
#  LICENSE file in the root directory of this source tree. An additional grant
#  of patent rights can be found in the PATENTS file in the same directory.
#
#
import __builtin__
import mock
from mock import mock_open
import pytest

old_open = __builtin__.open
__builtin__.open = mock_open(read_data='000003e8 00000040 000f4240 3b9aca00\n')

from atcd import AtcdThriftHandlerTask
from sparts.tests.base import SingleTaskTestCase

'''
class AtcdThriftHandlerTaskTest(SingleTaskTestCase):

    TASK = AtcdThriftHandlerTask

    def setUp(self):
        super(AtcdThriftHandlerTaskTest, self).setUp()

    def test_nothing(self):
        self.assertTrue(True)
'''
