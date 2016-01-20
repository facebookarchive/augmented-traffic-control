#!/usr/bin/env python
#
#  Copyright (c) 2014, Facebook, Inc.
#  All rights reserved.
#
#  This source code is licensed under the BSD-style license found in the
#  LICENSE file in the root directory of this source tree. An additional grant
#  of patent rights can be found in the PATENTS file in the same directory.
#
#
import functools
import logging
import logging.handlers
import os
import shlex
import socket
import subprocess
from subprocess import Popen
import time

from sparts.tasks.thrift import NBServerTask
from sparts.tasks.thrift import ThriftHandlerTask
from sparts.sparts import option

from atc_thrift import Atcd

# Atc thrift files
import atc_thrift.ttypes
from atc_thrift.ttypes import PacketCapture
from atc_thrift.ttypes import PacketCaptureException
from atc_thrift.ttypes import PacketCaptureFile
from atc_thrift.ttypes import ReturnCode
from atc_thrift.ttypes import TrafficControl
from atc_thrift.ttypes import TrafficControlException
from atc_thrift.ttypes import TrafficControlRc
from atc_thrift.ttypes import TrafficControlledDevice

from atcd.access_manager import AccessManager
from atcd.access_manager import AccessTokenException
from atcd.idmanager import IdManager
from atcd.AtcdDBQueueTask import AtcdDBQueueTask

from sqlite3 import OperationalError


def AccessCheck(func):
    @functools.wraps(func)
    def wrapper(self, tc_or_dev, *args, **kwargs):
        if isinstance(tc_or_dev, TrafficControl):
            dev = tc_or_dev.device
        elif isinstance(tc_or_dev, TrafficControlledDevice):
            dev = tc_or_dev
        else:
            raise TypeError(
                "You are using this decorator on the wrong kind of function."
                " Valid functions are `name(self, tc)` or `name(self, device)`"
            )
        if self.access_manager.access_allowed(dev):
            return func(self, tc_or_dev)
        else:
            raise TrafficControlException(
                code=ReturnCode.ACCESS_DENIED,
                message="The device {0} is not allowed to control {1}"
                .format(dev.controllingIP, dev.controlledIP),
            )

    return wrapper


def from_module_import_class(modulename, classname):
    """Import a class from a module.

    Allows importing a class from a module by using the class and module name
    as strings.
    ex: from_module_import_class('os.path', 'basename')

    Args:
        modulename: A string containing the name of the module to import from.
        classname: A string containing the name of the class to import.

    Returns:
        None

    Raises:
        AttributeError: An error accessing the class.
        ImportError: An error accessing the module.
    """
    klass = getattr(
        __import__(modulename, globals(), locals(), [classname], -1),
        classname
    )
    globals()[classname] = klass


class AtcdNBServerTask(NBServerTask):
    """Atcd Non Blocking Thrift server.

    Overrides sparks' Non blocking thrift server defaults for Atcd use.
    """
    DEFAULT_PORT = 9090
    DEFAULT_HOST = '127.0.0.1'


class AtcdThriftHandlerTask(ThriftHandlerTask):
    """Atcd's thrift handler.

    This is the main entry point of the program that implements the atcd.thrift
    interface definition.
    Platform specific behaviour will be implemented in Atcd`Platform`Shaper
    class.
    """
    ID_MANAGER_ID_MIN = 1
    ID_MANAGER_ID_MAX = 2**16

    MODULE = Atcd
    DEPS = [AtcdDBQueueTask]
    DEFAULT_LAN = 'eth1'
    DEFAULT_WAN = 'eth0'
    DEFAULT_IPTABLES = '/sbin/iptables'
    DEFAULT_TCPDUMP = '/usr/sbin/tcpdump'
    DEFAULT_PCAP_DIR = '/tmp'
    DEFAULT_PCAP_URL_BASE = 'http://localhost:80'
    DEFAULT_BURST_SIZE = 12000
    DEFAULT_MODE = 'secure'

    OPT_PREFIX = 'atcd'

    lan_name = option(
        default=DEFAULT_LAN,
        metavar='LAN',
        help='name of the LAN interface [%(default)s]',
        name='lan',
    )
    wan_name = option(
        default=DEFAULT_WAN,
        metavar='WAN',
        help='name of the WAN interface [%(default)s]',
        name='wan',
    )
    iptables = option(
        default=DEFAULT_IPTABLES,
        metavar='IPTABLES',
        help='location of the iptables binary [%(default)s]'
    )
    tcpdump = option(
        default=DEFAULT_TCPDUMP,
        metavar='TCPDUMP',
        help='location of the tcpdump binary [%(default)s]'
    )
    pcap_dir = option(
        default=DEFAULT_PCAP_DIR,
        metavar='PCAP_DIR',
        help='Directory to store pcap files [%(default)s]'
    )
    pcap_url_base = option(
        default=DEFAULT_PCAP_URL_BASE,
        metavar='PCAP_URL_BASE',
        help='URL for pcap service [%(default)s]'
    )
    burst_size = option(
        default=DEFAULT_BURST_SIZE,
        metavar='BURST_SIZE',
        type=int,
        help='Amount of bytes that can be burst at a capped speed '
                '[%(default)s]'
    )
    dont_drop_packets = option(
        action='store_true',
        help='[EXPERIMENTAL] Do not drop packets when going above max allowed'
             ' rate. Packets will be queued instead. Please mind that this'
             ' option will likely disappear in the future and is only provided'
             '  as a workaround until better longer term solution is found.',
    )
    fresh_start = option(
        action='store_true',
        help='Bypass saved shapings from a previous run [%(default)s]',
    )

    mode = option(
        choices=['secure', 'unsecure'],
        default=DEFAULT_MODE,
        help='In which mode should atcd run? [%(default)s]',
    )

    @staticmethod
    def factory():
        """Static method to discover and import the shaper to use.

        Discover the platform on which Atcd is running and import the shaping
        backend for this platform.

        Returns:
            The shaping backend class

        Raises:
            NotImplementedError: the shaping backend class couldn't be imported
        """
        os_name = os.uname()[0]
        klass = 'Atcd{0}Shaper'.format(os_name)
        # If not imported yet, try to import
        try:
            if klass not in globals():
                from_module_import_class(
                    'atcd.backends.{0}'.format(os_name.lower()), klass
                )
        except AttributeError:
            raise NotImplementedError('{0} is not implemented!'.format(klass))
        except ImportError:
            raise NotImplementedError(
                '{0} backend is not implemented!'.format(os_name.lower())
            )
        return globals()[klass]

    def initTask(self):
        """Thrift handler task initialization.

        Performs the steps needed to initialize the shaping subsystem.
        """
        super(AtcdThriftHandlerTask, self).initTask()

        # Do this first because it can error out and it's better to
        # error out before touching the networking stacks
        self.db_task = self.service.tasks.AtcdDBQueueTask

        self.lan = {'name': self.lan_name}
        self.wan = {'name': self.wan_name}
        self._links_lookup()

        self._ip_to_id_map = {}
        self._id_to_ip_map = {}
        self.initialize_id_manager()
        self.ip_to_pcap_proc_map = {}
        self.initialize_shaping_system()

        # Map of IP address to tc object that is currently
        # being used to shape traffic from that IP.
        # {ip: {'tc': tc, 'timeout': timeout}}
        # {'10.0.0.1': {'tc': TrafficControl(...), 'timeout': 1234567890.1234}}
        self._current_shapings = {}

        self.access_manager = AccessManager(secure=self.mode != 'unsecure')
        if not self.fresh_start:
            self.logger.info('Restoring shaped connection from DB')
            self._restore_saved_shapings()

    def _links_lookup(self):
        """Initialize our mapping from network interface name to their device
        id. Will raise and exception if one of the device is not found
        """
        raise NotImplementedError('Subclass should implement this')

    def initialize_id_manager(self):
        """Initialize the Id Manager. This is architecture dependant as the
        shaping subsystems may have different requirements.
        """
        self.idmanager = IdManager(
            first_id=type(self).ID_MANAGER_ID_MIN,
            max_id=type(self).ID_MANAGER_ID_MAX
        )

    def _restore_saved_shapings(self):
        """Restore the shapings from the sqlite3 db.
        """
        # Isolate the things we are using eval on to reduce possible clownyness
        # later on, also this way we don't have unused imports from importing
        # blindly for eval
        names = [
            'TrafficControlledDevice', 'TrafficControl', 'Shaping',
            'TrafficControlSetting', 'Loss', 'Delay', 'Corruption', 'Reorder'
        ]
        globals = {name: getattr(atc_thrift.ttypes, name) for name in names}

        # CurrentShapings(ip varchar primary key, tc blob, timeout int)
        results = []
        try:
            results = self.db_task.get_saved_shapings()
        except OperationalError:
            self.logger.exception('Unable to perform DB operation')
        for result in results:
            tc = eval(result['tc'], globals)
            timeout = float(result['timeout'])
            if timeout > time.time():
                tc.timeout = timeout - time.time()
                try:
                    self.startShaping(tc)
                except TrafficControlException as e:
                    # We have a shaping set in database that is denied
                    # probably because it was set in unsecure mode, passing
                    if (
                            e.code == ReturnCode.ACCESS_DENIED and
                            self.mode == 'secure'):
                        self.logger.warn(
                            'Shaping Denied in secure mode, passing:'
                            ' {0}'.format(e.message)
                        )
                        continue
                    raise
            else:
                self.db_task.queue.put(
                    (tc.device.controlledIP, 'remove_shaping')
                )

    def stop(self):
        """Implements sparts.vtask.VTask.stop()

        Each shaping platform should implement its own in order to clean
        its state before shutting down the main loop.
        """
        raise NotImplementedError('Subclass should implement this')

    def initialize_shaping_system(self):
        """Initialize the shaping subsystem.

        Each shaping platform should implement its own.
        """
        raise NotImplementedError('Subclass should implement this')

    def set_logger(self):
        """Initialize the logging subsystem.
        """
        self.logger = logging.getLogger(__name__)
        self.logger.setLevel(logging.DEBUG)
        fmt = logging.Formatter(fmt=logging.BASIC_FORMAT)
        # create console handler and set level to debug
        ch = logging.StreamHandler()
        ch.setLevel(logging.DEBUG)
        ch.setFormatter(fmt=fmt)
        self.logger.addHandler(ch)
        # create syslog handler and set level to debug
        sh = logging.handlers.SysLogHandler(address='/dev/log')
        sh.setLevel(logging.DEBUG)
        sh.setFormatter(fmt=fmt)
        self.logger.addHandler(sh)

    def getShapedDeviceCount(self):
        """Get the number of devices currently being shaped.

        Returns:
            The number of devices currently shaped.
        """
        self.logger.info("Request getShapedDeviceCount")
        return len(self._ip_to_id_map)

    @AccessCheck
    def startShaping(self, tc):
        """Start shaping a connection for a given device.

        Implements the `startShaping` thrift method.
        If the connection is already being shaped, the shaping will be updated
        and the old one deleted.

        Args:
            A TrafficControl object that contains the device to be shaped, the
            settings and the timeout.

        Returns:
            A TrafficControlRc object with code and message set to reflect
            success/failure.

        Raises:
            A TrafficControlException with code and message set on uncaught
            exception.
        """
        self.logger.info("Request startShaping {0}".format(tc))
        # Sanity checking
        # IP
        try:
            socket.inet_aton(tc.device.controlledIP)
        except Exception as e:
            return TrafficControlRc(
                code=ReturnCode.INVALID_IP,
                message="Invalid IP {}".format(tc.device.controlledIP))
        # timer
        if tc.timeout < 0:
            return TrafficControlRc(
                code=ReturnCode.INVALID_TIMEOUT,
                message="Invalid Timeout {}".format(tc.timeout))

        new_id = None
        try:
            new_id = self.idmanager.new()
        except Exception as e:
            return TrafficControlRc(
                code=ReturnCode.ID_EXHAUST,
                message="No more session available: {0}".format(e))

        old_id = self._ip_to_id_map.get(tc.device.controlledIP, None)
        old_settings = self._current_shapings.get(
            tc.device.controlledIP, {}
        ).get('tc')

        tcrc = self._shape_interface(
            new_id,
            self.wan,
            tc.device.controlledIP,
            tc.settings.up,
        )
        if tcrc.code != ReturnCode.OK:
            return tcrc
        tcrc = self._shape_interface(
            new_id,
            self.lan,
            tc.device.controlledIP,
            tc.settings.down,
        )
        # If we failed to set shaping for LAN interfaces, we should remove
        # the shaping we just created for the WAN
        if tcrc.code != ReturnCode.OK:
            self._unshape_interface(
                new_id,
                self.wan,
                tc.device.controlledIP,
                tc.settings.up
            )
            return tcrc
        self._add_mapping(new_id, tc)
        self.db_task.queue.put(((tc, tc.timeout + time.time()), 'add_shaping'))
        # if there were an existing id, remove it from dict
        if old_id is not None:
            self._unshape_interface(
                old_id,
                self.wan,
                tc.device.controlledIP,
                old_settings.settings.up,
            )
            self._unshape_interface(
                old_id,
                self.lan,
                tc.device.controlledIP,
                old_settings.settings.down,
            )
            del self._id_to_ip_map[old_id]
            self.idmanager.free(old_id)

        return TrafficControlRc(code=ReturnCode.OK)

    @AccessCheck
    def stopShaping(self, dev):
        """Stop shaping a connection for a given traffic controlled device.

        Implements the `stopShaping` thrift method.

        Args:
            A TrafficControlledDevice object that contains the shaped device.

        Returns:
            A TrafficControlRc object with code and message set to reflect
            success/failure.

        Raises:
            A TrafficControlException with code and message set on uncaught
            exception.
        """
        self.logger.info(
            "Request stopShaping for ip {0}".format(dev.controlledIP)
        )
        try:
            socket.inet_aton(dev.controlledIP)
        except Exception as e:
            return TrafficControlRc(
                code=ReturnCode.INVALID_IP,
                message="Invalid IP {0}: {1}".format(dev.controlledIP, e))

        id = self._ip_to_id_map.get(dev.controlledIP, None)
        shaping = self._current_shapings.get(dev.controlledIP, {}).get('tc')
        if id is not None:
            self._unshape_interface(
                id,
                self.wan,
                dev.controlledIP,
                shaping.settings.up,
            )
            self._unshape_interface(
                id,
                self.lan,
                dev.controlledIP,
                shaping.settings.down,
            )
            self._del_mapping(id, dev.controlledIP)
            self.db_task.queue.put((dev.controlledIP, 'remove_shaping'))
            self.idmanager.free(id)
        else:
            return TrafficControlRc(
                code=ReturnCode.UNKNOWN_SESSION,
                message="No session for IP {} found".format(dev.controlledIP))
        return TrafficControlRc(code=ReturnCode.OK)

    def _unshape_interface(self, mark, eth, ip, settings):
        """Unshape traffic for a given IP/setting on a network interface
        """
        raise NotImplementedError('Subclass should implement this')

    def _shape_interface(self, mark, eth, ip, shaping):
        """Shape traffic for a given IP
        """
        raise NotImplementedError('Subclass should implement this')

    def isShaped(self, dev):
        self.logger.info(
            "Request isShaped for ip {0}".format(dev.controlledIP)
        )
        return dev.controlledIP in self._ip_to_id_map

    def getCurrentShaping(self, dev):
        """Get the TrafficControl object used to shape a
            TrafficControlledDevice.

        Args:
            dev: a TrafficControlledDevice.

        Returns:
            A TrafficControl object representing the current shaping for the
            device.

        Raises:
            A TrafficControlException if there is no TC object for that IP
        """

        self.logger.info(
            'Request getCurrentShaping for ip {0}'.format(dev.controlledIP)
        )
        shaping = self._current_shapings.get(dev.controlledIP, {}).get('tc')
        if shaping is None:
            raise TrafficControlException(
                code=ReturnCode.UNKNOWN_IP,
                message='This IP ({0}) is not being shaped'.format(
                    dev.controlledIP
                )
            )
        return shaping

    def _add_mapping(self, id, tc):
        """Adds a mapping from id to IP address and vice versa.

        It also updates the dict mapping IPs to TrafficControl configs.

        Args:
            id: the id to map.
            tc: the TrafficControl object to map.
        """
        self._id_to_ip_map[id] = tc.device.controlledIP
        self._ip_to_id_map[tc.device.controlledIP] = id
        self._current_shapings[tc.device.controlledIP] = {
            'tc': tc, 'timeout': time.time() + tc.timeout}

    def _del_mapping(self, id, ip):
        """Removes mappings from IP to id and id to IP.

        Also  remove the mapping from IP to TrafficControl configs.
        """

        try:
            del self._id_to_ip_map[id]
            del self._ip_to_id_map[ip]
            del self._current_shapings[ip]
        except KeyError:
            self.logger.exception('Unable to remove key from dict')

    def run_cmd(self, cmd):
        self.logger.info("Running {}".format(cmd))
        return subprocess.call(shlex.split(cmd))

    def _pcap_filename(self, ip, start_time):
        return "%s_%d.cap" % (ip, start_time)

    def _pcap_parse_filename(self, filename):
        if filename.endswith(".cap"):
            ip, start_time = filename.replace(".cap", "").split("_")
            return ip, int(start_time)

    def _pcap_url(self, filename):
        return os.path.join(self.pcap_url_base, filename)

    def _pcap_full_path(self, filename):
        return os.path.join(self.pcap_dir, filename)

    def _pcap_file_size(self, filename):
        try:
            return int(os.path.getsize(self._pcap_full_path(filename)))
        except OSError:
            return 0

    def _cleanup_packet_capture_procs(self):
        '''Delete finished procs from the map'''
        for ip, p in self.ip_to_pcap_proc_map.items():
            if not p or p.poll() is not None:
                del self.ip_to_pcap_proc_map[ip]

    @AccessCheck
    def startPacketCapture(self, dev, timeout=3600):
        """Start a tcpdump process to capture packets for an ipaddr.

        The process will run until the timeout expires or stopPacketCapture()
        is called.

        Args:
            dev: a TrafficControlledDevice.
            timeout: int Max time for tcpdump process to run.

        Returns:
            True if process started ok, otherwise False.
        """
        self.logger.info(
            "Request startPacketCapture for ip {0}, timeout {1}".format(
                dev.controlledIP, timeout))
        start_time = time.time()
        filename = self._pcap_filename(dev.controlledIP, start_time)
        cmd = """timeout {timeout!s}
            {tcpdump} -vvv -s0 -i {eth} -w {filepath} host {ip}""".format(
            timeout=timeout,
            tcpdump=self.tcpdump,
            eth=self.lan["name"],
            filepath=self._pcap_full_path(filename),
            ip=dev.controlledIP)
        # Daemonize set the umask to 0o27 which prevents the http proxy service
        # from reading the file. For lack of better solution for now, we can
        # change the umask before spawning the subprocess and then restore its
        # original value
        umask = os.umask(0)
        p = Popen(shlex.split(cmd))
        os.umask(umask)
        if p and p.poll() is None:
            p.pcap = PacketCapture(
                ip=dev.controlledIP,
                start_time=start_time,
                file=PacketCaptureFile(
                    name=filename,
                    url=self._pcap_url(filename),
                    bytes=0),
                pid=p.pid)
            self.ip_to_pcap_proc_map[dev.controlledIP] = p
            return p.pcap
        else:
            raise PacketCaptureException(
                message="Failed to start tcpdump process")

    @AccessCheck
    def stopPacketCapture(self, dev):
        """Stop a tcpdump process that was started with startPacketCapture().

        Args:
           dev: a TrafficControlledDevice.

        Returns:
           The HTTP URL for the pcap file or empty string.
        """
        self.logger.info(
            "Request stopPacketCapture for ip {0}".format(dev.controlledIP)
        )
        self._cleanup_packet_capture_procs()
        if dev.controlledIP in self.ip_to_pcap_proc_map:
            p = self.ip_to_pcap_proc_map[dev.controlledIP]
            p.terminate()
            # Wait a few secs for processes to die, while cleaning up dead ones
            max_secs = 5
            start_time = time.time()
            while p.poll() is None and (time.time() - start_time) < max_secs:
                time.sleep(0.5)
            if p.poll() is None:
                p.kill()
            p.pcap.file.bytes = self._pcap_file_size(p.pcap.file.name)
            return p.pcap
        else:
            raise PacketCaptureException(
                message="No capture proc for given ipaddr")

    def stopAllPacketCaptures(self):
        """Stop all running tcpdump procs.
        """
        self.logger.info("Request stopAllPacketCaptures")
        self._cleanup_packet_capture_procs()
        if self.ip_to_pcap_proc_map:
            for p in self.ip_to_pcap_proc_map.values():
                p.terminate()
            # Wait a few secs for processes to die, while cleaning up dead ones
            max_secs = 5
            start_time = time.time()
            while self.ip_to_pcap_proc_map and \
                    (time.time() - start_time) < max_secs:
                time.sleep(0.5)
                self._cleanup_packet_capture_procs()
        if self.ip_to_pcap_proc_map:
            for p in self.ip_to_pcap_proc_map.values():
                p.kill()

    def listPacketCaptures(self, dev):
        """List the packet captures available for a given device.

        Args:
            dev: a TrafficControlledDevice.

        Returns:
            A list of PacketCapture ojbects.
        """
        ip = dev.controlledIP
        self.logger.info("Request listPacketCaptures for ip {0}".format(ip))
        pcap_list = []
        for filename in os.listdir(self.pcap_dir):
            if not filename.endswith(".cap"):
                continue
            file_ip, start_time = self._pcap_parse_filename(filename)
            if not file_ip == ip:
                continue
            pcap = PacketCapture(
                ip=ip,
                start_time=start_time,
                file=PacketCaptureFile(
                    name=filename,
                    url=self._pcap_url(filename),
                    bytes=self._pcap_file_size(filename)))
            pcap_list.append(pcap)
        return pcap_list

    def listRunningPacketCaptures(self):
        """List the running packet captures.

        Returns:
           A list of PacketCapture ojbects.
        """
        self.logger.info("Request listRunningPacketCaptures")
        pcap_list = []
        self._cleanup_packet_capture_procs()
        for ip, p in self.ip_to_pcap_proc_map.items():
            p.pcap.file.bytes = self._pcap_file_size(p.pcap.file.name)
            pcap_list.append(p.pcap)
        return pcap_list

    def stop_expired_shapings(self):
        """Stop shaping that have expired.
        """
        expired_devs = [
            attrs['tc'].device
            for ip, attrs in self._current_shapings.iteritems()
            if attrs['timeout'] <= time.time()
        ]
        for dev in expired_devs:
            self.logger.info('Shaping for Device "{0}" expired'.format(dev))
            self.logger.debug('calling stopShaping for "{0}"'.format(dev))
            self.stopShaping(dev)

    def requestToken(self, ip, duration):
        """Returns a unique, random access code.

        Random token to be given to a host to control the `ip`.
        The token validity is limited in time.

        Args:
            ip: The IP to control.
            duration: How long the token will be valid for.

        Returns:
            An AccessToken.
        """

        self.logger.info(
            "Request requestToken({0}, {1})".format(ip, duration)
        )
        token = self.access_manager.generate_token(ip, duration)
        return token

    def requestRemoteControl(self, dev, accessToken):
        """Request to control a remote device.

        Returns true if the token given is a valid token for the remote IP
            according to the totp object stored for that IP

        Args:
            dev: The TrafficControlledDevice.
            accessToken: The token to grant access.
        Returns:
            True if access is granted, False otherwise.
        """

        self.logger.info(
            "Request requestControl({0}, {1})".format(dev, accessToken)
        )
        access_granted = False
        try:
            self.access_manager.validate_token(
                dev,
                accessToken,
            )
            access_granted = True
        except AccessTokenException:
            self.logger.exception("Access Denied for request")
        return access_granted

    def getDevicesControlledBy(self, ip):
        """Get the devices controlled by a given IP.

        Args:
            ip: The IP of the controlling host.

        Returns:
            A list of RemoteControlInstance.
        """
        return self.access_manager.get_devices_controlled_by(ip)

    def getDevicesControlling(self, ip):
        """Get the devices controlling a given IP.

        Args:
            ip: The IP of the controlled host.

        Returns:
            A list of RemoteControlInstance.
        """
        return self.access_manager.get_devices_controlling(ip)
