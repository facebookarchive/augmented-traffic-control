#
#  Copyright (c) 2014, Facebook, Inc.
#  All rights reserved.
#
#  This source code is licensed under the BSD-style license found in the
#  LICENSE file in the root directory of this source tree. An additional grant
#  of patent rights can be found in the PATENTS file in the same directory.
#
#
from django.db import models


class Profile(models.Model):
    name = models.CharField(max_length=100, blank=False, null=False)
    content = models.CharField(max_length=1024, blank=False, null=False)
