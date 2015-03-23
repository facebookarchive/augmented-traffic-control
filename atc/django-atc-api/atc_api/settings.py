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
Settings for ATC_API are all namespaced in the ATC_API setting.
For example your project's `settings.py` file might look like this:

ATC_API = {
    'ATCD_HOST': 'localhost',
    'ATCD_PORT': 9090,
    'DEFAULT_TC_TIMEOUT': 24 * 60 * 60,
    'PROXY_IPS': ['127.0.0.1'],
}

This module provides the `atc_api_settings` object, that is used to access
ATC_API settings. It first check for user settings and then fall back on the
defaults.
'''

from django.conf import settings

USER_SETTINGS = getattr(settings, 'ATC_API', None)

DEFAULTS = {
    'ATCD_HOST': 'localhost',
    'ATCD_PORT': 9090,
    # Default timeout is a day in seconds
    'DEFAULT_TC_TIMEOUT': 24 * 60 * 60,
    'PROXY_IPS': ['127.0.0.1'],
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


atc_api_settings = APISettings(USER_SETTINGS, DEFAULTS)
