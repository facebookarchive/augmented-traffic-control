#!/usr/bin/env python
#
#  Copyright (c) 2014, Facebook, Inc.
#  All rights reserved.
#
#  This source code is licensed under the BSD-style license found in the
#  LICENSE file in the root directory of this source tree. An additional grant
#  of patent rights can be found in the PATENTS file in the same directory.
#
#

import threading


class IdManager(object):
    '''A class to manage disctributing ID objects'''
    def __init__(self, first_id=0, max_id=None):
        '''initialise the id manager class
        A minimun and maximum ID can be provided
        at initialisation time.'''
        self.first_id = first_id
        self.max_id = max_id
        self.next_available = first_id
        self.spares = set()
        self.lock = threading.Lock()

    def free(self, id):
        '''return an ID to the pool of available IDs'''
        with self.lock:
            if id == self.next_available - 1:
                self.next_available -= 1
            else:
                self.spares.add(id)

    def new(self):
        '''claim an ID from the pool of IDs, if no more IDs are available,
        throw an exception'''
        with self.lock:
            try:
                return self.spares.pop()
            except:
                next_avail = self.next_available
                if self.max_id is not None and \
                   self.next_available > self.max_id:
                    raise Exception(
                        "ID pool exhausted, max id is {0}".format(self.max_id)
                    )

                self.next_available += 1
                return next_avail
