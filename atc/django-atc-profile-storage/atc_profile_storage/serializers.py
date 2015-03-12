#
#  Copyright (c) 2014, Facebook, Inc.
#  All rights reserved.
#
#  This source code is licensed under the BSD-style license found in the
#  LICENSE file in the root directory of this source tree. An additional grant
#  of patent rights can be found in the PATENTS file in the same directory.
#
#
from atc_profile_storage.models import Profile
from rest_framework import serializers
import ast


class ProfileSerializer(serializers.ModelSerializer):
    class Meta:
        model = Profile
        fields = ('id', 'name', 'content')

    def to_representation(self, instance):
        ret = super(serializers.ModelSerializer, self).to_representation(
            instance)
        ret['content'] = ast.literal_eval(ret['content'])
        return ret
