#
#  Copyright (c) 2014, Facebook, Inc.
#  All rights reserved.
#
#  This source code is licensed under the BSD-style license found in the
#  LICENSE file in the root directory of this source tree. An additional grant
#  of patent rights can be found in the PATENTS file in the same directory.
#
#
from django.conf.urls import patterns
from django.conf.urls import url
from atc_api.views import AtcApi, AuthApi, TokenApi

urlpatterns = patterns(
    '',
    url('^shape/$', AtcApi.as_view()),
    url('^shape/'
        '(?P<address>[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3})/$',
        AtcApi.as_view()
        ),
    url('^auth/$', AuthApi.as_view()),
    url('^auth/'
        '(?P<address>[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3})/$',
        AuthApi.as_view()),
    url('^token/$', TokenApi.as_view()),
)
