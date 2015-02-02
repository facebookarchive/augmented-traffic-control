from subprocess import call, Popen, PIPE
import paramiko

from e2e.host import Host


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

    def up(self, *names):
        for name in names:
            if name not in self.vms:
                r = call(['vagrant', 'up', name])
                if r != 0:
                    raise RuntimeError('Could not bring up vagrant vm '+name)

    def destroy(self, name):
        r = call(['vagrant', 'destroy', '-f', name])
        if r != 0:
            raise RuntimeError('Could not destroy vagrant vm '+name)

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
                  stdin=None
                  )
        p.wait()
        if p.returncode != 0:
            raise RuntimeError('Could not get ssh-config for ' + repr(name))
        ssh_config = paramiko.SSHConfig()
        ssh_config.parse(p.stdout)
        p.stdout.close()
        return ssh_config.lookup(name)


Vagrant = _vagrant()


def setUpModule():
    # make sure these VMs are running.
    # defined in Vagrantfile
    pass  # Vagrant.up('client', 'server', 'gateway')


def tearDownModule():
    pass
