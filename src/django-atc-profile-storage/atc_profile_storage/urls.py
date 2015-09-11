#
#  Copyright (c) 2014, Facebook, Inc.
#  All rights reserved.
#
#  This source code is licensed under the BSD-style license found in the
#  LICENSE file in the root directory of this source tree. An additional grant
#  of patent rights can be found in the PATENTS file in the same directory.
#
#
from django.conf.urls import url
from atc_profile_storage import views

urlpatterns = [
    url(r'^$', views.profile_list),
    url(r'^(?P<pk>[0-9]+)/$', views.profile_detail),
]
