#
#  Copyright (c) 2014, Facebook, Inc.
#  All rights reserved.
#
#  This source code is licensed under the BSD-style license found in the
#  LICENSE file in the root directory of this source tree. An additional grant
#  of patent rights can be found in the PATENTS file in the same directory.
#
#

from speed import parseIPerfSpeed
import httplib
import json
import time


def speedBetween(client, server, time=30, udp=False):
    server_ip = server.getIp()

    srv_cmd = 'iperf -s' + (' -u' if udp else '') + ' -p 5001'
    cli_cmd = 'iperf -t ' + str(time) + (' -u' if udp else '') + \
        ' -c ' + server_ip + ' 5001'

    with server.proc(srv_cmd):
        s = client.cmd(cli_cmd)
        return parseIPerfSpeed(s.splitlines()[-1])


def shape(gateway, host, speed):
    '''
    Shape the provided host using the given gateway to the provided speed.

    If host is None, the gateway will use the HTTP request's remote IP.

    Gateway may be either a Host object or an IP address string.
    '''
    if isinstance(gateway, str):
        gw_ip = gateway
    else:
        gw_ip = gateway.getIp()

    shaping = {
        'down': {
            'rate': speed.kbps(),
            'loss': {
                'percentage': 0.0,
                'correlation': 0.0
            },
            'delay': {
                'delay': 0,
                'jitter': 0,
                'correlation': 0.0
            },
            'corruption': {
                'percentage': 0.0,
                'correlation': 0.0
            },
            'reorder': {
                'percentage': 0.0,
                'correlation': 0.0,
                'gap': 0
            },
            'iptables_options': []
        },
        'up': {
            'rate': speed.kbps(),
            'loss': {
                'percentage': 0.0,
                'correlation': 0.0
            },
            'delay': {
                'delay': 0,
                'jitter': 0,
                'correlation': 0.0
            },
            'corruption': {
                'percentage': 0.0,
                'correlation': 0.0
            },
            'reorder': {
                'percentage': 0.0,
                'correlation': 0.0,
                'gap': 0
            },
            'iptables_options': []
        }
    }

    url = '/api/v1/shape/'
    exc_f = 'Could not shape host: {}'
    if host is not None:
        shaped_ip = host.getIp()
        exc_f = 'Could not shape host %s: {}' % shaped_ip
        url += shaped_ip + '/'

    h = httplib.HTTPConnection(gw_ip, 8000, timeout=3)
    try:
        h.request(
            'POST',
            url,
            json.dumps(shaping),
            {'Content-Type': 'application/json'})
        r = h.getresponse()
        if r.status != httplib.CREATED:
            raise RuntimeError(
                exc_f.format(shaped_ip, r.status))
    finally:
        h.close()


def unshape(gateway, host):
    gw_ip = gateway.getIp()
    shaped_ip = host.getIp()

    h = httplib.HTTPConnection(gw_ip, 8000, timeout=3)
    try:
        h.request('DELETE', '/api/v1/shape/{}/'.format(shaped_ip))
        r = h.getresponse()
        if r.status != httplib.NO_CONTENT:
            raise RuntimeError(
                'Could not shape host {}: {}'.format(shaped_ip, r.status))
    finally:
        h.close()

    # Atcd takes some time to start affecting traffic.
    time.sleep(3.0)
