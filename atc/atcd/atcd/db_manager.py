#
#  Copyright (c) 2014, Facebook, Inc.
#  All rights reserved.
#
#  This source code is licensed under the BSD-style license found in the
#  LICENSE file in the root directory of this source tree. An additional grant
#  of patent rights can be found in the PATENTS file in the same directory.
#
#
import logging
import sqlite3


class SQLiteManager(object):
    """ Manage various SQLite operations for ATCd
    """

    SHAPING_INSERT_QUERY = \
        'INSERT OR REPLACE INTO CurrentShapings values (?, ?, ?)'
    SHAPING_CREATE_QUERY = \
        'CREATE TABLE IF NOT EXISTS CurrentShapings('\
        'ip VARCHAR PRIMARY KEY NOT NULL, tc_obj BLOB, timeout INT)'
    SHAPING_TABLE_NAME = 'CurrentShapings'
    SHAPING_IP_COL = 0
    SHAPING_TC_COL = 1
    SHAPING_TIMOUT_COL = 2

    def __init__(self, file_name, logger=None):
        self.logger = logger or logging.getLogger()
        self.file_name = file_name
        with self._get_conn() as conn:
            conn.execute(SQLiteManager.SHAPING_CREATE_QUERY)
        conn.close()

    def get_saved_shapings(self):
        """ Querys the db and returns a list of the
            TrafficControl objects that are stored there.
            returns as a list of dicts that have a key for 'tc' and 'timeout'
        """
        query = 'SELECT * FROM CurrentShapings'
        with self._get_conn() as conn:
            results = conn.execute(query).fetchall()
        conn.close()
        # shapings = [{'tc': tc_obj, 'timeout': 123456}, ... ]
        shapings = []
        for result in results:
            shapings.append(
                {
                    'tc': result[SQLiteManager.SHAPING_TC_COL],
                    'timeout': result[SQLiteManager.SHAPING_TIMOUT_COL]
                }
            )
        return shapings

    def add_shaping(self, tc, timeout):
        with self._get_conn() as conn:
            conn.execute(
                SQLiteManager.SHAPING_INSERT_QUERY,
                (tc.device.controlledIP, repr(tc), timeout)
            )
        conn.close()

    def remove_shaping(self, ip):
        query = 'DELETE FROM CurrentShapings WHERE ip = ?'
        with self._get_conn() as conn:
            conn.execute(query, (ip,))
        conn.close()

    def _get_conn(self):
        try:
            conn = sqlite3.connect(self.file_name)
        except sqlite3.OperationalError:
            self.logger.error(
                'Unable to access db file: {0}'.format(self.file_name)
            )
            raise
        return conn
