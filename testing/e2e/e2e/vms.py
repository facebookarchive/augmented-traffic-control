from e2e.speed import parseIPerfSpeed


def speedBetween(client, server):
    server_ip = server.getIp()

    with server.proc('iperf -s -p 5001'):
        s = client.cmd('iperf -c ' + server_ip + ' 5001')
        return parseIPerfSpeed(s.splitlines()[-1])


def shape(gateway, host, speed):
    pass


def unshape(gateway, host):
    pass
