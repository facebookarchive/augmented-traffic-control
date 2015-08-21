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
from atc_api.utils import get_client_ip
from atc_thrift.ttypes import TrafficControlException, TrafficControl
from atc_thrift.ttypes import TrafficControlledDevice, AccessToken

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
            data=request.data,
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
        setting_serializer = SettingSerializer(data=request.data)
        device_serializer = DeviceSerializer(
            data=request.data,
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
            data=request.data,
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


class AuthApi(APIView):

    @serviced
    def get(self, request, service, address=None):
        '''
        Returns the addresses that the provided address is allowed to shape.
        '''
        if address is None:
            address = get_client_ip(request)

        controlled_ips = []

        for addr in service.getDevicesControlledBy(address):
            if addr is None:
                break
            controlled_ips.append({
                'controlled_ip': addr.device.controlledIP,
                'valid_until': addr.timeout,
            })

        data = {
            'address': address,
            'controlled_ips': controlled_ips,
        }
        return Response(data, status=status.HTTP_200_OK)

    @serviced
    def post(self, request, service, address=None):
        '''
        Authorizes one address to shape another address,
        based on the provided auth token.
        '''
        if address is None:
            return Response(
                {'details': 'no address provided'},
                status=status.HTTP_400_BAD_REQUEST
                )
        controlled_ip = address

        controlling_ip = get_client_ip(request)

        if 'token' not in request.data:
            token = None
        else:
            token = AccessToken(token=request.data['token'])

        dev = TrafficControlledDevice(
            controlledIP=controlled_ip,
            controllingIP=controlling_ip
            )

        worked = service.requestRemoteControl(dev, token)

        if not worked:
            return Response(
                {'details': 'invalid token provided'},
                status=status.HTTP_401_UNAUTHORIZED,
                )

        print 'Worked:', worked

        data = {
            'controlling_ip': controlling_ip,
            'controlled_ip': controlled_ip,
        }

        return Response(data, status=status.HTTP_200_OK)


class TokenApi(APIView):

    @serviced
    def get(self, request, service):
        '''
        Returns the current authorization token for the provided address.
        '''
        # default duration...
        # 3 days in seconds
        duration = 3 * 24 * 60 * 60

        if 'duration' in request.query_params:
            duration = int(request.query_params['duration'])

        address = get_client_ip(request)

        stuff = service.requestToken(address, duration)

        data = {
            'token': stuff.token,
            'interval': stuff.interval,
            'valid_until': stuff.valid_until,
            'address': address,
        }

        return Response(data, status=status.HTTP_200_OK)
