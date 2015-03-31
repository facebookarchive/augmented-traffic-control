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
import pytest
import time
import unittest

from atc_thrift.ttypes import TrafficControlledDevice
from atcd.access_manager import AccessManager
from atc_thrift.ttypes import AccessToken
from atcd.access_manager import AccessTokenException
from atcd.access_manager import AtcdTOTP
from mock import Mock

INTERVAL = 60


@pytest.fixture
def control_allowed():
    return {
        ('1.1.1.1', '2.2.2.1'): 20,
        ('1.1.1.2', '2.2.2.2'): 5,
        ('1.1.1.1', '2.2.2.4'): 15,
        ('1.1.1.1', '2.2.2.5'): 5,
        ('1.1.1.3', '2.2.2.1'): 5,
        ('1.1.1.4', '2.2.2.1'): 15,
    }


@pytest.fixture
def ip_to_totp_map():
    return {
        '2.2.2.1': {
            'totp': AtcdTOTP(s='12345', interval=60),
            'duration': 15,
        },
        '2.2.2.2': {
            'totp': AtcdTOTP(s='12345', interval=60),
            'duration': 5,
        },
    }


@pytest.fixture
def am():
    return AccessManager()


@pytest.fixture
def fake_am(am, control_allowed, ip_to_totp_map):
    am._control_allowed = control_allowed
    am._ip_to_totp_map = ip_to_totp_map
    return am


@pytest.fixture
def fail_verify(monkeypatch):
    monkeypatch.setattr(AtcdTOTP, 'verify', Mock(return_value=False))


@pytest.fixture
def succeed_verify(monkeypatch):
    monkeypatch.setattr(AtcdTOTP, 'verify', Mock(return_value=True))


def _make_device(controlling, controlled=None):
    return TrafficControlledDevice(
        controllingIP=controlling,
        controlledIP=controlled
    )


def _make_token(token):
    return AccessToken(token=token)


class TestAtcdTOTP(unittest.TestCase):

    interval = 30
    s = 'wrn3pqx5uqxqvnqr'

    def test_valid_until(self):
        t = 1297553958
        endtime30s = 1297553970
        endtime10s = 1297553960
        with Timecop(t):
            totp = AtcdTOTP(interval=30, s=self.s)
            dt = datetime.datetime.fromtimestamp(t)
            self.assertEqual(
                datetime.datetime.fromtimestamp(endtime30s),
                totp.valid_until(dt)
            )
            totp = AtcdTOTP(interval=10, s=self.s)
            dt = datetime.datetime.fromtimestamp(t)
            self.assertEqual(
                datetime.datetime.fromtimestamp(endtime10s),
                totp.valid_until(dt)
            )
        assert True


class TestAccessManager():

    def setup_method(self, method):

        def mocktime():
            return 10
        self._old_time = time.time
        time.time = mocktime

    def teardown_method(self, method):
        time.time = self._old_time

    def test_generate_token(self, fake_am):
        l = len(fake_am._ip_to_totp_map.keys())
        fake_am.generate_token('1.1.1.1', 10)
        assert len(fake_am._ip_to_totp_map.keys()) == l+1

        fake_am.generate_token('1.1.1.1', 30)
        assert len(fake_am._ip_to_totp_map.keys()) == l+1

    def test_controlled_by_existing(self, fake_am):
        controlling_by = fake_am.get_devices_controlled_by('1.1.1.1')
        assert len(controlling_by) == 2

    def test_controlled_by_non_existent(self, fake_am):
        controlling_by = fake_am.get_devices_controlled_by('3.3.3.3')
        assert len(controlling_by) == 0

    def test_controlling_existing(self, fake_am):
        controlling_by = fake_am.get_devices_controlling('2.2.2.1')
        assert len(controlling_by) == 2

    def test_controlling_non_existent(self, fake_am):
        controlling_by = fake_am.get_devices_controlling('3.3.3.3')
        assert len(controlling_by) == 0

    def test_access_allowed_controlling_ip_none(self, fake_am):
        # controllingIP = None
        assert not fake_am.access_allowed(_make_device(None, '2.2.2.5'))
        # Allowed in non-secure mode
        fake_am.secure = False
        assert fake_am.access_allowed(_make_device(None, '2.2.2.5'))

    def test_access_allowed_valid(self, fake_am):
        # valid entry
        dev = TrafficControlledDevice(
            controllingIP='1.1.1.1',
            controlledIP='2.2.2.1'
        )
        assert fake_am.access_allowed(dev)

    def test_access_allowed_non_existent(self, fake_am):
        # entry does not exist
        dev = TrafficControlledDevice(
            controllingIP='1.1.1.1',
            controlledIP='2.2.2.2'
        )
        assert not fake_am.access_allowed(dev)
        # Allowed in non-secure mode
        fake_am.secure = False
        assert fake_am.access_allowed(dev)

    def test_access_allowed_expired(self, fake_am):
        # expired entry
        dev = TrafficControlledDevice(
            controllingIP='1.1.1.1',
            controlledIP='2.2.2.5'
        )
        assert not fake_am.access_allowed(dev)
        # Allowed in non-secure mode
        fake_am.secure = False
        assert fake_am.access_allowed(dev)

    def test_access_allowed_self(self, fake_am):
        # expired entry
        dev = TrafficControlledDevice(
            controllingIP='1.1.1.1',
            controlledIP='1.1.1.1'
        )
        assert fake_am.access_allowed(dev)

    def test_validate_token_valid(self, fake_am, succeed_verify):
        fake_am.validate_token(
            _make_device('1.1.1.1', '2.2.2.1'),
            _make_token('12345'),
        )

    def test_validate_token_invalid(self, fake_am, fail_verify):
        with pytest.raises(AccessTokenException) as excinfo:
            fake_am.validate_token(
                _make_device('1.1.1.1', '2.2.2.1'),
                _make_token('12344'),
            )
        assert str(excinfo.value) == 'Access denied for device pair'

    # FIXME, this is not really handling expiration properly
    def test_validate_token_expired_valid(self, fake_am, fail_verify):
        with pytest.raises(AccessTokenException) as excinfo:
            fake_am.validate_token(
                _make_device('1.1.1.2', '2.2.2.2'),
                _make_token('12345'),
            )
        assert str(excinfo.value) == 'Access denied for device pair'

    # FIXME, this is not really handling expiration properly
    def test_validate_token_expired_invalid(self, fake_am, fail_verify):
        with pytest.raises(AccessTokenException) as excinfo:
            fake_am.validate_token(
                _make_device('1.1.1.2', '2.2.2.2'),
                _make_token('12344'),
            )
        assert str(excinfo.value) == 'Access denied for device pair'

    def test_validate_token_non_existent(self, fake_am):
        with pytest.raises(AccessTokenException) as excinfo:
            fake_am.validate_token(
                _make_device('1.1.1.2', '2.2.2.0'),
                _make_token('12344'),
            )
        assert str(excinfo.value) == \
            '''That remote device hasn't generated a code yet'''


# Directly copied from https://github.com/nathforge/pyotp/blob/master/test.py
class Timecop(object):
    """
    Half-assed clone of timecop.rb, just enough to pass our tests.
    """

    def __init__(self, freeze_timestamp):
        self.freeze_timestamp = freeze_timestamp

    def __enter__(self):
        self.real_datetime = datetime.datetime
        datetime.datetime = self.frozen_datetime()

    def __exit__(self, type, value, traceback):
        datetime.datetime = self.real_datetime

    def frozen_datetime(self):
        class FrozenDateTime(datetime.datetime):
            @classmethod
            def now(cls):
                return cls.fromtimestamp(timecop.freeze_timestamp)

        timecop = self
        return FrozenDateTime
