#
#  Copyright (c) 2014, Facebook, Inc.
#  All rights reserved.
#
#  This source code is licensed under the BSD-style license found in the
#  LICENSE file in the root directory of this source tree. An additional grant
#  of patent rights can be found in the PATENTS file in the same directory.
#
#
import datetime
import pyotp
import time

from atc_thrift.ttypes import AccessToken
from atc_thrift.ttypes import TrafficControlledDevice
from atc_thrift.ttypes import RemoteControlInstance


def _dev_to_tuple(device):
    return device.controllingIP, device.controlledIP


def _tuple_to_dev(tup):
    return TrafficControlledDevice(
        controllingIP=tup[0],
        controlledIP=tup[1],
    )


def _remote_control_instance(tup, timeout):
    return RemoteControlInstance(
        device=_tuple_to_dev(tup),
        timeout=timeout,
    )


class AccessTokenException(Exception):
    pass


class AtcdTOTP(pyotp.TOTP):
    def valid_until(self, for_time):
        """
        Returns the time that a code will expire, given a Time object.

        @param [Time] Time object
        @return [Time] time the code that would be generated at `for_time`
        is valid until
        """
        valid_time = (self.timecode(for_time) + 1) * self.interval
        valid_datetime = datetime.datetime.fromtimestamp(valid_time)
        return valid_datetime


class AccessManager(object):
    ACCESS_TOKEN_INTERVAL = 60

    def __init__(self, secure=True):
        self._ip_to_totp_map = {}
        self._control_allowed = {}
        self.secure = secure

    def generate_token(self, ip, duration):
        """
        takes an ip to generate an AccessToken for and a duration that the
        remote device will be granted control of the ip once the token is used
        """
        totp_dict = self._ip_to_totp_map.get(ip)
        if totp_dict is None:
            # Timeout changed to 60 seconds from the default 30 as it may take
            # more than 30 sec to get the code, go to other client and enter it
            totp = AtcdTOTP(
                interval=self.ACCESS_TOKEN_INTERVAL,
                s=pyotp.random_base32()
            )
            self._ip_to_totp_map[ip] = {
                'totp': totp,
                'duration': duration
            }
        else:
            totp = totp_dict.get('totp')
            if duration != totp_dict.get('duration'):
                totp_dict['duration'] = duration
                self._ip_to_totp_map[ip] = totp_dict

        timestamp = datetime.datetime.now()

        return AccessToken(
            token=totp.at(timestamp),
            interval=self.ACCESS_TOKEN_INTERVAL,
            # valid_until returns time as a datetime.datetime object
            # this converts it to a float time
            valid_until=time.mktime(totp.valid_until(timestamp).timetuple())
        )

    def validate_token(self, dev, access_token):
        """
        takes a TrafficControlDevice and an AccessToken and if that device and
        token are a valid combo, stores the time dev.controllingIP has access
            internally for lookup later.
        This either returns None on success or
            raises an AccessTokenException on failure
        """
        # Shortcuts
        # Of course you can control yourself!
        if not (dev.controllingIP == dev.controlledIP):
            totp_dict = self._ip_to_totp_map.get(dev.controlledIP, {})
            totp = totp_dict.get('totp')
            duration = totp_dict.get('duration')
            if not (totp and duration):
                raise AccessTokenException("That remote device hasn't"
                                           " generated a code yet")

            if totp.verify(access_token.token):
                timeout = time.time() + duration
                self._control_allowed[_dev_to_tuple(dev)] = timeout
            else:
                raise AccessTokenException("Access denied for device pair")

    def access_allowed(self, dev):
        """
        Decides whether or not dev.controllingIP has access to control
        dev.controlledIP
        @returns boolean
        """
        # Non secure mode, access granted everytime
        if not self.secure:
            return True

        if dev.controllingIP == dev.controlledIP:
            return True
        dev_tuple = _dev_to_tuple(dev)
        timeout = self._control_allowed.get(dev_tuple)
        if timeout:
            if timeout > time.time():
                return True
            else:
                del self._control_allowed[dev_tuple]
        return False

    def get_devices_controlled_by(self, ip):
        '''
        Implementation for atcd.getDevicesControlledBy
        '''
        now = time.time()

        def is_valid(key, val):
            return key[0] == ip and val > now

        return [
            _remote_control_instance(key, val)
            for (key, val) in self._control_allowed.items()
            if is_valid(key, val)
        ]

    def get_devices_controlling(self, ip):
        '''
        Implementation for atcd.getDevicesControlling
        '''
        now = time.time()

        def is_valid(key, val):
            return key[1] == ip and val > now

        return [
            _remote_control_instance(key, val)
            for (key, val) in self._control_allowed.items()
            if is_valid(key, val)
        ]
