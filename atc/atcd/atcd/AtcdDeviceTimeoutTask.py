#
#  Copyright (c) 2014, Facebook, Inc.
#  All rights reserved.
#
#  This source code is licensed under the BSD-style license found in the
#  LICENSE file in the root directory of this source tree. An additional grant
#  of patent rights can be found in the PATENTS file in the same directory.
#
#
from atcd.AtcdThriftHandlerTask import AtcdThriftHandlerTask
from sparts.tasks.periodic import PeriodicTask


class AtcdDeviceTimeoutTask(PeriodicTask):
    INTERVAL = 10.0

    def initTask(self):
        super(AtcdDeviceTimeoutTask, self).initTask()
        self.required_task = AtcdThriftHandlerTask.factory()

    def execute(self):
        AtcdMainTask = self.service.requireTask(
            self.required_task
        )
        AtcdMainTask.stop_expired_shapings()
