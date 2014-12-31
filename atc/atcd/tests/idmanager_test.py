#
#  Copyright (c) 2014, Facebook, Inc.
#  All rights reserved.
#
#  This source code is licensed under the BSD-style license found in the
#  LICENSE file in the root directory of this source tree. An additional grant
#  of patent rights can be found in the PATENTS file in the same directory.
#
#
from atcd import idmanager
import unittest


class TestIdManager(unittest.TestCase):
    MAXID = 10

    def setUp(self):
        self.idm = idmanager.IdManager(1, TestIdManager.MAXID)

    def test_assignement(self):
        nbiters = 5

        idx = self.idm.new()
        self.assertEqual(idx, 1)

        for i in range(nbiters):
            idx = self.idm.new()
        self.assertEqual(idx, nbiters + 1)

        # return the id we allocated last
        self.idm.free(nbiters+1)
        # return id 2 and 5
        self.idm.free(2)
        self.idm.free(5)

        s = set()
        s.add(self.idm.new())
        s.add(self.idm.new())
        self.assertEqual(s, set([2, 5]))

    def test_exhaustion(self):
        idx = 0
        # test that we throw an exception
        with self.assertRaises(Exception):
            for i in xrange(TestIdManager.MAXID + 1):
                idx = self.idm.new()
        self.assertEqual(idx, TestIdManager.MAXID)
