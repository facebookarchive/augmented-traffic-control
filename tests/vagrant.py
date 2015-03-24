#
#  Copyright (c) 2014, Facebook, Inc.
#  All rights reserved.
#
#  This source code is licensed under the BSD-style license found in the
#  LICENSE file in the root directory of this source tree. An additional grant
#  of patent rights can be found in the PATENTS file in the same directory.
#
#

from subprocess import Popen, PIPE
import paramiko

from host import Host


# For use with the 'with' python feature.
class _sshGroup(object):

    @classmethod
    def closeAll(cls, clients):
        if len(clients) == 0:
            return
        try:
            clients[0].close()
        finally:
            cls.closeAll(clients[1:])

    def __init__(self):
        self.clients = []

    def append(self, client):
        self.clients.append(client)

    def __enter__(self):
        return tuple(self.clients)

    def __exit__(self, type, value, tb):
        _sshGroup.closeAll(self.clients)
        return False


class _vagrant(object):
    vms = []

    def ssh(self, *names):
        clients = _sshGroup()
        for name in names:
            ssh_config = self.sshConfig(name)
            clients.append(Host(ssh_config))
        return clients

    def sshConfig(self, name):
        p = Popen(['vagrant', 'ssh-config', name],
                  stdout=PIPE,
                  stderr=None,
                  stdin=None,
                  cwd='tests/',
                  )
        p.wait()
        if p.returncode != 0:
            raise RuntimeError('Could not get ssh-config for ' + repr(name))
        ssh_config = paramiko.SSHConfig()
        ssh_config.parse(p.stdout)
        p.stdout.close()
        return ssh_config.lookup(name)


Vagrant = _vagrant()
