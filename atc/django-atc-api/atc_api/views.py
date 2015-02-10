#
#  Copyright (c) 2014, Facebook, Inc.
#  All rights reserved.
#
#  This source code is licensed under the BSD-style license found in the
#  LICENSE file in the root directory of this source tree. An additional grant
#  of patent rights can be found in the PATENTS file in the same directory.
#
#
from atc_api.atcd_client import atcdClient
from atc_api.serializers import SettingSerializer, DeviceSerializer
from atc_api.settings import atc_api_settings
from atc_thrift.ttypes import TrafficControlException
from atc_thrift.ttypes import TrafficControl

from functools import wraps
from rest_framework.exceptions import APIException
from rest_framework.exceptions import ParseError
from rest_framework.response import Response
from rest_framework.views import APIView
from rest_framework import status


class BadGateway(APIException):
    status_code = 502
    default_detail = 'Could not connect to ATC gateway.'


def serviced(method):
    '''
    A decorator to check if the service is available or not.
    Raise a BadGateway exception on failure to connect to the atc gateway
    '''
    @wraps(method)
    def decorator(cls, request, *args, **kwargs):
        service = atcdClient()
        if service is None:
            raise BadGateway()
        return method(cls, request, service, *args, **kwargs)
    return decorator


class AtcApi(APIView):
    '''
    If `address` is not provided, we default to the client IP or forwarded IP
    '''

    @serviced
    def get(self, request, service, address=None, format=None):
        ''''
        Get the current shaping for an IP. If address is None, defaults to
        the client IP
        @return the current shaping applied or 404 if the IP is not being
        shaped
        '''
        device_serializer = DeviceSerializer(
            data=request.DATA,
            context={'request': request, 'address': address},
        )
        if not device_serializer.is_valid():
            raise ParseError(detail=device_serializer.errors)

        dev = device_serializer.save()
        try:
            tc = service.getCurrentShaping(dev)
        except TrafficControlException as e:
            return Response(
                {'detail': e.message},
                status=status.HTTP_404_NOT_FOUND,
            )

        serializer = SettingSerializer(tc.settings)
        return Response(
            serializer.data,
            status=status.HTTP_200_OK
        )

    @serviced
    def post(self, request, service, address=None, format=None):
        ''''
        Set shaping for an IP. If address is None, defaults to
        the client IP
        @return the profile that was set on success
        '''
        setting_serializer = SettingSerializer(data=request.DATA)
        device_serializer = DeviceSerializer(
            data=request.DATA,
            context={'request': request, 'address': address},
        )
        if not setting_serializer.is_valid():
            raise ParseError(detail=setting_serializer.errors)

        if not device_serializer.is_valid():
            raise ParseError(detail=device_serializer.errors)

        setting = setting_serializer.save()
        device = device_serializer.save()

        tc = TrafficControl(
            device=device,
            settings=setting,
            timeout=atc_api_settings.DEFAULT_TC_TIMEOUT,
        )

        try:
            tcrc = service.startShaping(tc)
        except TrafficControlException as e:
            return Response(e.message, status=status.HTTP_401_UNAUTHORIZED)
        result = {'result': tcrc.code, 'message': tcrc.message}
        if tcrc.code:
            raise ParseError(detail=result)

        return Response(
            setting_serializer.data,
            status=status.HTTP_201_CREATED
        )

    @serviced
    def delete(self, request, service, address=None, format=None):
        '''
        Delete the shaping for an IP, if no IP is specified, default to the
        client IP
        '''
        device_serializer = DeviceSerializer(
            data=request.DATA,
            context={'request': request, 'address': address},
        )
        if not device_serializer.is_valid():
            return Response(
                device_serializer.errors,
                status=status.HTTP_400_BAD_REQUEST,
            )

        device = device_serializer.save()

        try:
            tcrc = service.stopShaping(device)
        except TrafficControlException as e:
            return Response(e.message, status=status.HTTP_401_UNAUTHORIZED)

        result = {'result': tcrc.code, 'message': tcrc.message}
        if tcrc.code:
            raise ParseError(detail=result)
        return Response(status=status.HTTP_204_NO_CONTENT)
