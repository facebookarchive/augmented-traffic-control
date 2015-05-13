#
#  Copyright (c) 2014, Facebook, Inc.
#  All rights reserved.
#
#  This source code is licensed under the BSD-style license found in the
#  LICENSE file in the root directory of this source tree. An additional grant
#  of patent rights can be found in the PATENTS file in the same directory.
#
#
from atc_api.utils import get_client_ip
from atc_thrift.ttypes import Corruption
from atc_thrift.ttypes import Delay
from atc_thrift.ttypes import Loss
from atc_thrift.ttypes import Reorder
from atc_thrift.ttypes import Shaping
from atc_thrift.ttypes import TrafficControlledDevice
from atc_thrift.ttypes import TrafficControlSetting

from rest_framework import serializers
from thrift.Thrift import TType

import socket


def validate_ipaddr(ipaddr):
    try:
        socket.inet_aton(ipaddr)
        return True
    except socket.error:
        return False


class ThriftSerializer(serializers.Serializer):
    # Should be set by the serializer to the concrete thrift class
    # to be serialized.
    _THRIFT_CLASS = None

    # A map of renamed fields.
    # Keys in the map are the names of thrift fields. Their values
    # are the names of the serializer fields they correspond to.
    _THRIFT_RENAMED_FIELDS = {}

    def create(self, attrs):
        args = {}

        for field_tuple in self._THRIFT_CLASS.thrift_spec:
            if not field_tuple:
                continue

            _, thrift_type, arg_name, _, default = field_tuple

            f_name = arg_name
            if arg_name in self._THRIFT_RENAMED_FIELDS:
                f_name = self._THRIFT_RENAMED_FIELDS[arg_name]

            serializer = self.fields[f_name]

            if f_name not in attrs:
                args[arg_name] = default
                continue

            if thrift_type == TType.STRUCT:
                args[arg_name] = serializer.create(attrs[f_name])
            else:
                # Primitive
                args[arg_name] = attrs[f_name]

        return self._THRIFT_CLASS(**args)


class BaseShapingSettingSerializer(ThriftSerializer):
    percentage = serializers.FloatField(default=0)
    correlation = serializers.FloatField(default=0)


class DelaySerializer(ThriftSerializer):
    _THRIFT_CLASS = Delay

    delay = serializers.IntegerField(default=0)
    jitter = serializers.IntegerField(default=0)
    correlation = serializers.FloatField(default=0)


class LossSerializer(BaseShapingSettingSerializer):
    _THRIFT_CLASS = Loss


class CorruptionSerializer(BaseShapingSettingSerializer):
    _THRIFT_CLASS = Corruption


class ReorderSerializer(BaseShapingSettingSerializer):
    _THRIFT_CLASS = Reorder

    gap = serializers.IntegerField(default=0)


class IptablesOptionsField(serializers.Field):

    def to_representation(self, data):
        if isinstance(data, list):
            return data
        else:
            msg = self.error_messages['invalid']
            raise serializers.ValidationError(msg)

    def to_internal_value(self, obj):
        if obj:
            return obj
        else:
            return []


class ShapingSerializer(ThriftSerializer):
    _THRIFT_CLASS = Shaping

    rate = serializers.IntegerField(default=0, allow_null=True, required=False)
    loss = LossSerializer(default=None, allow_null=True, required=False)
    delay = DelaySerializer(default=None, allow_null=True, required=False)
    corruption = CorruptionSerializer(
        default=None, allow_null=True, required=False)
    reorder = ReorderSerializer(default=None, allow_null=True, required=False)
    iptables_options = IptablesOptionsField(
        default=None, allow_null=True, required=False)


class SettingSerializer(ThriftSerializer):
    _THRIFT_CLASS = TrafficControlSetting

    down = ShapingSerializer()
    up = ShapingSerializer()


class DeviceSerializer(ThriftSerializer):
    _THRIFT_CLASS = TrafficControlledDevice
    _THRIFT_RENAMED_FIELDS = {
        'controllingIP': 'client',
        'controlledIP': 'address'
    }

    address = serializers.CharField(
        max_length=16,
        allow_blank=True,
        allow_null=True,
        default=None,
        required=False
    )
    client = serializers.CharField(
        max_length=16,
        allow_blank=True,
        allow_null=True,
        default=None,
        required=False
    )

    def validate_address(self, value):
        # 'address' is optional, if not specified, we default to the
        # querying IP
        # `address` can be specified in 2 places: the URL or within the payload
        # The payload has priority and will be accessible through `value`
        # The value passed in the URL is accessible through the context

        if value is None or (isinstance(value, str) and len(value) == 0):
            if self.context.get('address'):
                value = self.context['address']
            else:
                value = self._get_client_ip()
        if not validate_ipaddr(value):
            raise serializers.ValidationError("Invalid IP address")
        return value

    def validate_client(self, value):
        # 'client' should not be provided by the payload.
        # It should always be the client IP as we get it from _get_client_ip()
        # This is merely here so we can use the serializer.
        return self._get_client_ip()

    def _get_client_ip(self):
        return get_client_ip(self.context['request'])
