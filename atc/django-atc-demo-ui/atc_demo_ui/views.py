#
#  Copyright (c) 2014, Facebook, Inc.
#  All rights reserved.
#
#  This source code is licensed under the BSD-style license found in the
#  LICENSE file in the root directory of this source tree. An additional grant
#  of patent rights can be found in the PATENTS file in the same directory.
#
#
from atc_demo_ui.settings import atc_demo_ui_settings

from django.shortcuts import render_to_response
from django.template import RequestContext


def index(request):
    context = {'atc_demo_ui_settings': atc_demo_ui_settings}
    return render_to_response(
        'atc_demo_ui/index.html',
        context,
        RequestContext(request)
    )
