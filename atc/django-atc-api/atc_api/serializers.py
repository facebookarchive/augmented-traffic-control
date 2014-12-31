#
#  Copyright (c) 2014, Facebook, Inc.
#  All rights reserved.
#
#  This source code is licensed under the BSD-style license found in the
#  LICENSE file in the root directory of this source tree. An additional grant
#  of patent rights can be found in the PATENTS file in the same directory.
#
#
from atc_thrift.ttypes import Corruption
from atc_thrift.ttypes import Delay
from atc_thrift.ttypes import Loss
from atc_thrift.ttypes import Reorder
from atc_thrift.ttypes import Shaping
from atc_thrift.ttypes import TrafficControlledDevice
from atc_thrift.ttypes import TrafficControlSetting

from rest_framework import serializers

import socket


def validate_ipaddr(ipaddr):
    try:
        socket.inet_aton(ipaddr)
        return True
    except socket.error:
        return False


class BaseShapingSettingSerializer(serializers.Serializer):
    percentage = serializers.FloatField(default=0)
    correlation = serializers.FloatField(default=0)


class DelaySerializer(serializers.Serializer):
    delay = serializers.IntegerField(default=0)
    jitter = serializers.IntegerField(default=0)
    correlation = serializers.FloatField(default=0)

    def restore_object(self, attrs, instance=None):
        return Delay(**attrs)


class LossSerializer(BaseShapingSettingSerializer):

    def restore_object(self, attrs, instance=None):
        return Loss(**attrs)


class CorruptionSerializer(BaseShapingSettingSerializer):

    def restore_object(self, attrs, instance=None):
        return Corruption(**attrs)


class ReorderSerializer(BaseShapingSettingSerializer):
    gap = serializers.IntegerField(default=0)

    def restore_object(self, attrs, instance=None):
        return Reorder(**attrs)


class IptablesOptionsField(serializers.WritableField):

    def from_native(self, data):
        if isinstance(data, list):
            return data
        else:
            msg = self.error_messages['invalid']
            raise serializers.ValidationError(msg)

    def to_native(self, obj):
        if obj:
            return obj
        else:
            return []


class ShapingSerializer(serializers.Serializer):
    rate = serializers.IntegerField(default=0, required=False)
    loss = LossSerializer(required=False)
    delay = DelaySerializer(required=False)
    corruption = CorruptionSerializer(required=False)
    reorder = ReorderSerializer(required=False)
    iptables_options = IptablesOptionsField(
        required=False
    )

    def restore_object(self, attrs, instance=None):
        return Shaping(**attrs)


class SettingSerializer(serializers.Serializer):
    down = ShapingSerializer()
    up = ShapingSerializer()

    def restore_object(self, attrs, instance=None):
        return TrafficControlSetting(**attrs)


class DeviceSerializer(serializers.Serializer):
    address = serializers.CharField(max_length=16, required=False)

    def validate_address(self, attrs, source):
        value = attrs.get(source, None)
        # 'address' is optional, if not specified, we default to the
        # querying IP
        if value is None:
            return attrs
        if not validate_ipaddr(value):
            raise serializers.ValidationError("Invalid IP address")
        return attrs

    def restore_object(self, attrs, instance=None):
        return self._make_device(
            self._get_address(attrs)
        )

    def _get_address(self, attrs):
        '''
            First we try to get the `address` from the context (URL),
            then from the json payload (attrs)
            and finally we default to the client IP
        '''
        return (
            self.context.get('address') or
            attrs.get('address', self._get_client_ip())
        )

    def _get_client_ip(self):
        '''Return the real IP of a client even when using a proxy'''
        request = self.context['request']
        if 'HTTP_X_REAL_IP' in request.META:
            return request.META['HTTP_X_REAL_IP']
        else:
            return request.META['REMOTE_ADDR']

    def _make_device(self, address):
        return TrafficControlledDevice(
            controlledIP=address,
            # FIXME: re-enable this once auth is fully operational
            # controllingIP=_get_client_ip(request)
        )
