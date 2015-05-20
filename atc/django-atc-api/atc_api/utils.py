#
#  Copyright (c) 2015, Facebook, Inc.
#  All rights reserved.
#
#  This source code is licensed under the BSD-style license found in the
#  LICENSE file in the root directory of this source tree. An additional grant
#  of patent rights can be found in the PATENTS file in the same directory.
#
#

from atc_api.settings import atc_api_settings


def get_client_ip(request):
    '''
    Return the real IP of a client even when using a proxy
    '''
    if 'HTTP_X_REAL_IP' in request.META:
        if request.META['REMOTE_ADDR'] not in atc_api_settings.PROXY_IPS:
            raise ValueError('HTTP_X_REAL_IP set by non-proxy')
        return request.META['HTTP_X_REAL_IP']
    else:
        return request.META['REMOTE_ADDR']
