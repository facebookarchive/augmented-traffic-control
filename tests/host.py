#
#  Copyright (c) 2014, Facebook, Inc.
#  All rights reserved.
#
#  This source code is licensed under the BSD-style license found in the
#  LICENSE file in the root directory of this source tree. An additional grant
#  of patent rights can be found in the PATENTS file in the same directory.
#
#

import paramiko
import logging

# size of receive buffers for stdout/stderr
BUFFER_SIZE = 1024 * 1024

# reduce paramiko logging spew
logging.getLogger('paramiko').setLevel(logging.WARNING)


# A class which does nothing
class Nothing(object):

    def __getattr__(self, name):
        def noop(*args, **kwargs):
            pass
        return noop


class Host(paramiko.SSHClient):
    name = ''
    client = None
    config = None

    def __init__(self, ssh_config):
        self.config = ssh_config
        self.name = ssh_config['hostname']
        self.client = paramiko.SSHClient()
        self.client.set_missing_host_key_policy(Nothing())

        self.client.connect(
            hostname=self.name,
            port=int(ssh_config['port']),
            username=ssh_config['user'],
            key_filename=ssh_config['identityfile'],
            )

    def getIp(self, prefix='192.168.'):
        # ip addr | grep 'inet' | awk '{print $2}' \
        #    | awk -F/ '{print $1}' | fgrep '192.168.' | head -n1
        out = self.cmd('ip addr')
        for line in out.splitlines():
            line = line.strip()
            if line.startswith('inet'):
                ip = line.split()[1].split('/')[0]
                if ip.startswith(prefix):
                    return ip
        raise RuntimeError('Could not determine ip of host ' + repr(self.name))

    def cmd(self, command):
        '''Runs the command over ssh. Returns stdout as a string'''
        return self.client.exec_command(command)[1].read()

    def proc(self, command):
        '''Runs the command in a pty.
        Returns a Process object representing the server-side process'''
        return Process(self, command)

    def close(self):
        self.client.close()

    def __exit__(self, type, val, tb):
        self.close()
        return False

    def __str__(self):
        return 'Host<' + self.name + '>'


class Process(object):
    commandName = ''
    channel = None
    host = None

    def __init__(self, host, command):
        self.host = host

        xport = self.host.client.get_transport()
        self.channel = xport.open_session()
        self.commandName = command.strip().split()[0]

        self.channel.exec_command(command)

    def stdout(self):
        return self.channel.recv(BUFFER_SIZE)

    def kill(self):
        self.channel.close()
        # FIXME: channel.close() will close the socket,
        # but not kill the remote process.
        # This is a simple workaround, but won't work in all cases.
        self.host.cmd('killall ' + self.commandName)

    def __enter__(self):
        return self

    def __exit__(self, type, val, tb):
        self.kill()
        return False
