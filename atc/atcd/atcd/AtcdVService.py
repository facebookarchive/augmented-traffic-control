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

from sparts.vservice import VService


class AtcdVService(VService):
    def initLogging(self):
        super(AtcdVService, self).initLogging()
        sh = logging.handlers.SysLogHandler(address='/dev/log')
        sh.setLevel(logging.DEBUG)
        self.logger.addHandler(sh)
        # Make sparts.tasks logging go to /var/log
        sparts_tasks_logger = logging.getLogger('sparts.tasks')
        sparts_tasks_logger.addHandler(sh)
