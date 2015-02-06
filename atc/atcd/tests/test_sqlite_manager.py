#
#  Copyright (c) 2014, Facebook, Inc.
#  All rights reserved.
#
#  This source code is licensed under the BSD-style license found in the
#  LICENSE file in the root directory of this source tree. An additional grant
#  of patent rights can be found in the PATENTS file in the same directory.
#
#
import pytest
import sqlite3
import tempfile
import time

import atc_thrift.ttypes
from atc_thrift.ttypes import Delay
from atc_thrift.ttypes import Shaping
from atc_thrift.ttypes import TrafficControl
from atc_thrift.ttypes import TrafficControlSetting
from atc_thrift.ttypes import TrafficControlledDevice
from atcd.db_manager import SQLiteManager

test_ipaddr = '10.01.10.01'


@pytest.fixture
def atc_db_file():
    """return a NamedTemporyFile (tempfile.NamedTemportFile) for use
    with testing ATC's SQLite DB features
    """
    return tempfile.NamedTemporaryFile(
        suffix='.db',
        prefix='atc_',
    )


@pytest.fixture
def dbm(atc_db_file):
    return SQLiteManager(atc_db_file.name)


@pytest.fixture
def test_shaping():
    return TrafficControl(
        device=TrafficControlledDevice(
            controlledIP=test_ipaddr
        ),
        timeout=86400,
        settings=TrafficControlSetting(
            down=Shaping(
                delay=Delay(
                    delay=197,
                ),
                rate=81,
            ),
            up=Shaping(
                delay=Delay(
                    delay=197,
                ),
                rate=81,
            )
        )
    )


@pytest.fixture
def test_db(dbm, test_shaping):
    dbm.add_shaping(test_shaping, time.time() + test_shaping.timeout)
    return dbm


class TestSQLiteManager():

    test_query = 'select ip,tc_obj,timeout from {0} where ip=?'.format(
        SQLiteManager.SHAPING_TABLE_NAME
    )

    def test_sqlite_file_not_found(self):
        with pytest.raises(sqlite3.OperationalError):
            SQLiteManager('/this/path/should/not/exist')

    def test_sqlite_init(self, dbm):
        test_conn = sqlite3.connect(dbm.file_name)
        sql = test_conn.execute(
            "select sql from sqlite_master where type='table' and name=?",
            (SQLiteManager.SHAPING_TABLE_NAME,)
        ).fetchone()
        assert SQLiteManager.SHAPING_CREATE_QUERY.replace(
            'IF NOT EXISTS ', ''
        ) in sql

    def test_sqlite_add_shaping(self, dbm, test_shaping):
        dbm.add_shaping(test_shaping, time.time() + test_shaping.timeout)
        test_conn = sqlite3.connect(dbm.file_name)
        results = test_conn.execute(self.test_query, (test_ipaddr,)).fetchone()
        assert results[SQLiteManager.SHAPING_TC_COL] == repr(test_shaping)
        names = [
            'TrafficControlledDevice', 'TrafficControl', 'Shaping',
            'TrafficControlSetting', 'Loss', 'Delay', 'Corruption', 'Reorder'
        ]
        globals = {name: getattr(atc_thrift.ttypes, name) for name in names}
        tc = eval(results[SQLiteManager.SHAPING_TC_COL], globals)
        assert tc == test_shaping

    def test_sqlite_remove_shaping(self, test_db):
        test_conn = sqlite3.connect(test_db.file_name)
        results = test_conn.execute(self.test_query, (test_ipaddr,)).fetchone()
        assert results
        # results = (ip, tc_obj, timeout)
        test_db.remove_shaping(results[0])
        results = test_conn.execute(self.test_query, (test_ipaddr,)).fetchone()
        assert not results

    def test_sqlite_get_saved_shapings(self, test_db):
        results = test_db.get_saved_shapings()
        assert len(results) > 0
        for result in results:
            assert isinstance(result, dict)
            assert 'tc' in result
            assert 'timeout' in result
        with sqlite3.connect(test_db.file_name) as test_conn:
            test_conn.execute(
                'DELETE FROM {0}'.format(SQLiteManager.SHAPING_TABLE_NAME)
            )
        results = test_db.get_saved_shapings()
        assert results == []
