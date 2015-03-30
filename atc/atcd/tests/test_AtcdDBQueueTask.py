#
#  Copyright (c) 2014, Facebook, Inc.
#  All rights reserved.
#
#  This source code is licensed under the BSD-style license found in the
#  LICENSE file in the root directory of this source tree. An additional grant
#  of patent rights can be found in the PATENTS file in the same directory.
#
#
from atc_thrift.ttypes import TrafficControl
from atc_thrift.ttypes import TrafficControlledDevice
from atcd.AtcdDBQueueTask import AtcdDBQueueTask
from sparts.sparts import option
from sparts.tests.base import SingleTaskTestCase

import mock
import pytest
import sqlite3
import tempfile


@pytest.fixture
def atc_db_file():
    """return a NamedTemporyFile (tempfile.NamedTemportFile) for use
    with testing ATC's SQLite DB features
    """
    return tempfile.NamedTemporaryFile(
        suffix='.db',
        prefix='atc_',
    )


class AtcdDBQueueTestTask(AtcdDBQueueTask):
    sqlite_file = option(default=atc_db_file().name)


class TestAtcdDBQueueTask(SingleTaskTestCase):

    TASK = AtcdDBQueueTestTask

    def test_get_saved_shaping(self):
        self.assertEqual(len(self.task.get_saved_shapings()), 0)

    def test_execute_missing_arg(self):
        # We silently discard the query...
        l = self.task.get_saved_shapings()
        self.task.execute(('obj',), 'context_unused')
        # DB is not modified, we expect the same content
        self.assertEqual(l, self.task.get_saved_shapings())

    def test_operational_error(self):
        # When there is a operational error, we just swallow the exception
        tc = self._make_tc_device()
        l = self.task.get_saved_shapings()
        with mock.patch('atcd.db_manager.sqlite3.connect') as mock_connect:
            mock_connect.side_effect = sqlite3.OperationalError('Op Error')
            self.task.execute(((tc, 10), 'add_shaping'), 'context_unused')
        self.assertEqual(l, self.task.get_saved_shapings())

    def test_execute_unkown_action(self):
        # unknown action is expected to raise an AttributeError exception
        with pytest.raises(AttributeError):
            self.task.execute(('obj', 'unknown_action'), 'context_unused')

    def test_add_shaping_wrong_arguments(self):
        # We expect a tc object, not a string.
        with pytest.raises(AttributeError):
            self.task.execute(
                (('tc', 'timeout'), 'add_shaping'),
                'context_unused'
            )

    def test_add_shaping_missing_arguments(self):
        # This should raise a TypeError exception.
        # add_shaping expects 2 arguments.
        with pytest.raises(TypeError):
            self.task.execute(
                ('tc', 'add_shaping'), 'context_unused'
            )

    def test_add_shaping_correct_arguments(self):
        # Test adding/removing a shaped device.
        ip = '1.1.1.1'
        tc = self._make_tc_device(ip=ip)
        self.task.execute(((tc, 10), 'add_shaping'), 'context_unused')
        self.assertEqual(len(self.task.get_saved_shapings()), 1)
        self.task.execute((ip, 'remove_shaping'), 'context_unused')
        self.assertEqual(len(self.task.get_saved_shapings()), 0)

    def test_remove_shaping_not_in_db(self):
        # When removing somehting not in DB, the number of entries
        # stay the same.
        ip = '1.1.1.1'
        self.assertEqual(len(self.task.get_saved_shapings()), 0)
        self.task.execute((ip, 'remove_shaping'), 'context_unused')
        self.assertEqual(len(self.task.get_saved_shapings()), 0)

    def _make_tc_device(self, ip='1.1.1.1'):
        tc = TrafficControl()
        tc.device = TrafficControlledDevice(ip, ip)
        return tc
