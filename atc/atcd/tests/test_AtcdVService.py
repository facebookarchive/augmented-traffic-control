#
#  Copyright (c) 2014, Facebook, Inc.
#  All rights reserved.
#
#  This source code is licensed under the BSD-style license found in the
#  LICENSE file in the root directory of this source tree. An additional grant
#  of patent rights can be found in the PATENTS file in the same directory.
#
#
from atcd.AtcdVService import AtcdVService
from sparts.tests.base import ServiceTestCase

import logging


class AtcdVServiceTest(ServiceTestCase):

    def getServiceClass(self):
        return AtcdVService

    def test_logger_use_syslog(self):
        handlers = self.service.logger.handlers
        self.assertTrue(
            'SysLogHandler' in [type(h).__name__ for h in handlers]
        )

    def test_logger_spart_syslog(self):
        handlers = logging.getLogger('sparts.tasks').handlers
        self.assertTrue(
            'SysLogHandler' in [type(h).__name__ for h in handlers]
        )
