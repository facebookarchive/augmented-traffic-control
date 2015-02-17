from e2e.speed import parseIPerfSpeed


def speedBetween(client, server):
    server_ip = server.getIp()

    with server.proc('iperf -s -p 5001'):
        s = client.cmd('iperf -c ' + server_ip + ' 5001')
        return parseIPerfSpeed(s.splitlines()[-1])


def shape(gateway, host, speed):
    '''
    curl -i http://192.168.20.2:8000/api/v1/shape/ -d '{"down":{"rate":10000,"loss":{"percentage":0.0,"correlation":0.0},"delay":{"delay":0,"jitter":0,"correlation":0.0},"corruption":{"percentage":0.0,"correlation":0.0},"reorder":{"percentage":0.0,"correlation":0.0,"gap":0},"iptables_options":[]},"up":{"rate":10000,"loss":{"percentage":0.0,"correlation":0.0},"delay":{"delay":0,"jitter":0,"correlation":0.0},"corruption":{"percentage":0.0,"correlation":0.0},"reorder":{"percentage":0.0,"correlation":0.0,"gap":0},"iptables_options":[]}}'
    '''
    pass


def unshape(gateway, host):
    pass
