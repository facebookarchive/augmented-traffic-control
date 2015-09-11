#
#  Copyright (c) 2014, Facebook, Inc.
#  All rights reserved.
#
#  This source code is licensed under the BSD-style license found in the
#  LICENSE file in the root directory of this source tree. An additional grant
#  of patent rights can be found in the PATENTS file in the same directory.
#
#
'''
Settings for ATC_DEMO_UI are all namespaced in the ATC_DEMO_UI setting.
For example your project's `settings.py` file might look like this:

ATC_DEMO_UI = {
    'SHORT_TITLE': 'ATC Demo UI',
    'TITLE': 'Augmented Traffic Control Demo UI',
    'EMAIL': 'atc@example.com',
    'INFO_MESSAGE': '',
    'REST_ENDPOINT': '/api/v1/',
    'BOOTSTRAP_VERSION': '3.3.0',
}

This module provides the `atc_demo_ui_settings` object, that is used to access
ATC_DEMO_UI settings. It first check for user settings and then fall back on
the defaults.
'''

from django.conf import settings

USER_SETTINGS = getattr(settings, 'ATC_DEMO_UI', None)

DEFAULTS = {
    'SHORT_TITLE': 'ATC Demo UI',
    'TITLE': 'Augmented Traffic Control Demo UI',
    'EMAIL': 'atc@example.com',
    'INFO_MESSAGE': '',
    'REST_ENDPOINT': '/api/v1/',
    'BOOTSTRAP_VERSION': '3.3.0',
}


class APISettings(object):

    def __init__(self, user_settings=None, defaults=None):
        self.__user_settings = user_settings or {}
        self.__defaults = defaults or {}

    def __getattr__(self, attr):
        if attr not in self.__defaults.keys():
            raise AttributeError("Invalid API setting: '%s'" % attr)

        try:
            # Check if user have set that key.
            val = self.__user_settings[attr]
        except KeyError:
            # Use defaults otherwise.
            val = self.__defaults[attr]

        return val


atc_demo_ui_settings = APISettings(USER_SETTINGS, DEFAULTS)
