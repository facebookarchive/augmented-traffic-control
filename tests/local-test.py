#!/usr/bin/env python

from subprocess import Popen, PIPE
import argparse

from speed import Speed, parseIPerfSpeed, KILOBITS
from vms import shape


ATC = '192.168.20.2'
IPERF = '192.168.10.10'
RATES = [
    Speed(10000, KILOBITS),
    Speed(1000, KILOBITS),
    Speed(100, KILOBITS),
    Speed(50, KILOBITS),
    Speed(25, KILOBITS),
    Speed(10, KILOBITS),
    Speed(5, KILOBITS),
]
NTESTS = 10


# fixme: move to test utils
def parseIPerfPrefix(s):
    b = s.find('0.0-')
    e = s.find(' sec', b)+4
    interval = s[b:e].split('-')[1]

    b = e + 4
    transfer = ' '.join(s[b:].split()[:2])

    return interval, transfer


# fixme: move to test utils
def parseIPerfOutput(stdout):
    line = stdout.splitlines()[-1].strip()
    interval, transfer = parseIPerfPrefix(line)
    speed = parseIPerfSpeed(line)
    return interval, transfer, speed


def run_iperf(server):
    p = Popen(['iperf', '-c', server],
              stdout=PIPE,
              stderr=None,
              stdin=None,
              )
    p.wait()
    if p.returncode != 0:
        raise RuntimeError('Could not run iperf')
    stdout = p.stdout.read()
    return parseIPerfOutput(stdout)


def test_network(gateway, server, rate):
    shape(gateway, None, rate)
    print_results(rate, run_iperf(server))


def print_header():
    print 'Shaping         Interval        Transfer        Bandwidth'


def print_results(rate, things):
    interval, transfer, speed = (things)
    print '%s\t%s\t%s\t%s' % (rate, interval, transfer, speed)


def rateList(rate_str):
    results = []
    for r in rate_str.split(','):
        results.append(Speed(int(r), KILOBITS))
    return results


def getGateway(server):
    p = Popen(['ip', 'route', 'get', server],
              stdout=PIPE,
              stderr=None,
              stdin=None,
              )
    p.wait()
    if p.returncode != 0:
        raise RuntimeError('Could not get route to server')
    stdout = p.stdout.read()
    return stdout.split()[2]


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('-r', '--rates',
                        default=rateList('10000,1000,100,50,25,10,5'),
                        type=rateList,
                        help='comma-separated list of rates in Kbps')
    parser.add_argument('-n', '--count',
                        type=int, default=10,
                        help='number of runs for each rate')
    parser.add_argument('--atc', default=None,
                        type=str,
                        help='IP address of ATC gateway')
    parser.add_argument('iperf',
                        type=str, help='IP address of iperf server')
    args = parser.parse_args()

    iperf = args.iperf
    if args.atc is None:
        atc = getGateway(iperf)
    else:
        atc = args.atc

    print_header()
    for rate in args.rates:
        for i in range(args.count):
            test_network(atc, iperf, rate)


if __name__ == '__main__':
    main()
