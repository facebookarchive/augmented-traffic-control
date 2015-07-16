#
#  Copyright (c) 2014, Facebook, Inc.
#  All rights reserved.
#
#  This source code is licensed under the BSD-style license found in the
#  LICENSE file in the root directory of this source tree. An additional grant
#  of patent rights can be found in the PATENTS file in the same directory.
#
#
import sys

from atcd.AtcdThriftHandlerTask import AtcdThriftHandlerTask

# Pyroute stuff
from pyroute2 import IPRoute
from pyroute2.netlink.rtnl import TC_H_ROOT
from pyroute2.netlink.rtnl import RTM_NEWTCLASS
from pyroute2.netlink.rtnl import RTM_DELTCLASS
from pyroute2.netlink.rtnl import RTM_NEWQDISC
from pyroute2.netlink.rtnl import RTM_DELQDISC
from pyroute2.netlink.rtnl import RTM_NEWTFILTER
from pyroute2.netlink.rtnl import RTM_DELTFILTER
from pyroute2.netlink import NetlinkError

# Atc thrift files
from atc_thrift.ttypes import ReturnCode
from atc_thrift.ttypes import TrafficControlRc


ETH_P_IP = 0x0800
PRIO = 1
HANDLE_MIN = 2
HANDLE_MAX = (2 ** 16) - 1


def int_to_classid(i):
    s = "{0:X}:{1:X}".format(i >> 16, 0xff & i)
    return s


class AtcdLinuxShaper(AtcdThriftHandlerTask):

    ID_MANAGER_ID_MIN = HANDLE_MIN
    ID_MANAGER_ID_MAX = HANDLE_MAX

    def initTask(self):
        self.ipr = IPRoute()
        super(AtcdLinuxShaper, self).initTask()

    def stop(self):
        self._release_ipr()

    def _release_ipr(self):
        self.ipr.close()

    def _links_lookup(self):
        try:
            self.lan['id'] = self.ipr.link_lookup(ifname=self.lan_name)[0]
            self.wan['id'] = self.ipr.link_lookup(ifname=self.wan_name)[0]
        except IndexError:
            self._release_ipr()
            msg = 'One of the following interfaces does not exist:' \
                ' {0}, {1}'.format(self.lan_name, self.wan_name)
            self.logger.critical(msg)
            raise Exception(msg)

    def initialize_shaping_system(self):
        """Initialize Iptables and TC subsystems
        Only call once as this will FLUSH all current
        shapings...
        """
        self.logger.info("Calling initialize_shaping_system")
        self._initialize_iptables()
        self._initialize_tc()

    def _initialize_iptables(self):
        """Initialize IPTables by flushing all rules in FORWARD chain
        from mangle table.
        """
        cmd = "{0} -t mangle -F FORWARD".format(self.iptables)
        self.run_cmd(cmd)

    def _initialize_tc_for_interface(self, eth):
        """Initialize TC on a given interface.

        If an exception is thrown, it will be forwarded to the main loop
        unless it can be ignored.

        Args:
            eth: the interface to flush TC on.

        Raises:
            NetlinkError: An error occured initializing TC subsystem.
            Exception: Any other exception thrown during initialization.
        """
        idx = 0x10000
        eth_name = eth['name']
        eth_id = eth['id']
        try:
            self.logger.info("deleting root QDisc on {0}".format(eth_name))
            self.ipr.tc(RTM_DELQDISC, None, eth_id, 0, parent=TC_H_ROOT)
        except Exception as e:
            # a (2, 'No such file or directory') can be thrown if there is
            # nothing to delete. Ignore such error, return the error otherwise
            if isinstance(e, NetlinkError) and e.code == 2:
                self.logger.warning(
                    "could not delete root QDisc. There might "
                    "have been nothing to delete")
            else:
                self.logger.exception(
                    'Initializing root Qdisc for {0}'.format(eth_name)
                )
                raise

        try:
            self.logger.info("setting root qdisc on {0}".format(eth_name))
            self.ipr.tc(RTM_NEWQDISC, "htb", eth_id, idx, default=0)
        except Exception as e:
            self.logger.exception(
                'Setting root Qdisc for {0}'.format(eth_name)
            )
            raise

        return TrafficControlRc(code=ReturnCode.OK)

    def _initialize_tc(self):
        """Initialize TC root qdisc on both LAN and WAN interface.
        """
        for netif in [self.lan, self.wan]:
            self._initialize_tc_for_interface(netif)

    def _unset_htb_class(self, mark, eth):
        """Given a mark and an interface, unset the HTB class.

        Args:
            mark: The mark based on which we delete the class.
            eth: The interface on which to delete that class id.

        Returns:
            A TrafficControlRc containing information on success/failure.
        """
        ifid = eth['id']
        idx = 0x10000 + mark
        try:
            self.logger.info(
                "deleting class on IFID {0}, classid {1}".format(
                    eth['name'], int_to_classid(idx)
                )
            )
            self.ipr.tc(RTM_DELTCLASS, 'htb', ifid, idx)
        except NetlinkError as e:
            return TrafficControlRc(
                code=ReturnCode.NETLINK_HTB_ERROR,
                message=str(e))
        except Exception as e:
            self.logger.exception('_unset_htb_class')
            exc_info = sys.exc_info()
            return TrafficControlRc(
                code=ReturnCode.UNKNOWN_HTB_ERROR,
                message=str(exc_info))

        return TrafficControlRc(code=ReturnCode.OK)

    def _set_htb_class(self, mark, eth, shaping):
        """Given a mark, an interface and shaping settings, set the HTB class.

        Args:
            mark: The mark based on which we create the class
            eth: The interface on which to create that class id.
            shaping: The shaping settings to set.

        Returns:
            A TrafficControlRc containing information on success/failure.
        """
        ifid = eth['id']
        idx = 0x10000 + mark
        parent = 0x10000
        self.logger.info(
            "create new HTB class on IFID {0}, classid {1},"
            "parent {2}, rate {3}kbits".format(
                eth['name'], int_to_classid(idx),
                int_to_classid(parent), shaping.rate or 2**22 - 1)
        )
        try:
            self.ipr.tc(
                RTM_NEWTCLASS, 'htb', ifid, idx,
                parent=parent,
                rate="{}kbit".format(shaping.rate or (2**22 - 1)),
            )
        except NetlinkError as e:
            return TrafficControlRc(
                code=ReturnCode.NETLINK_HTB_ERROR,
                message=str(e))
        except Exception as e:
            self.logger.exception('_set_htb_class')
            exc_info = sys.exc_info()
            return TrafficControlRc(
                code=ReturnCode.UNKNOWN_HTB_ERROR,
                message=str(exc_info))

        return TrafficControlRc(code=ReturnCode.OK)

    def _unset_netem_qdisc(self, mark, eth):
        """This is not needed as deleting the HTB class is sufficient
        to remove the netem qdisc"""
        pass

    def _set_netem_qdisc(self, mark, eth, shaping):
        """Given a mark, interface and shaping settings, create the NetEm
        Qdisc.

        Args:
            mark: The mark based on which we create the Qdisc.
            eth: The interface on which we will create the Qdisc.
            shaping: The shaping settings for that interface.

        Returns:
            A TrafficControlRc containing information on success/failure.
        """
        ifid = eth['id']
        parent = 0x10000 + mark
        idx = 0  # automatically assign a handleid
        self.logger.info(
            "create new Netem qdisc on IFID {0}, parent {1},"
            " loss {2}%, delay {3}".format(
                eth['name'], int_to_classid(parent),
                shaping.loss.percentage,
                shaping.delay.delay * 1000)
        )
        try:
            self.ipr.tc(
                RTM_NEWQDISC, 'netem', ifid, idx,
                parent=parent,
                loss=shaping.loss.percentage,
                delay=shaping.delay.delay * 1000,
                jitter=shaping.delay.jitter * 1000,
                delay_corr=shaping.delay.correlation,
                loss_corr=shaping.loss.correlation,
                prob_reorder=shaping.reorder.percentage,
                corr_reorder=shaping.reorder.correlation,
                gap=shaping.reorder.gap,
                prob_corrupt=shaping.corruption.percentage,
                corr_corrupt=shaping.corruption.correlation,
            )
        except NetlinkError as e:
            return TrafficControlRc(
                code=ReturnCode.NETLINK_NETEM_ERROR,
                message=str(e))
        except Exception as e:
            self.logger.exception('_set_netem_qdisc')
            exc_info = sys.exc_info()
            return TrafficControlRc(
                code=ReturnCode.UNKNOWN_NETEM_ERROR,
                message=str(exc_info))

        return TrafficControlRc(code=ReturnCode.OK)

    def _unset_filter(self, mark, eth):
        """Given a mark and an interface, delete the filter.

        Args:
            mark: The mark based on which we delete the filter.
            eth: The interface on which we delete the filter.

        Returns:
            A TrafficControlRc containing information on success/failure.
        """
        ifid = eth['id']
        parent = 0x10000
        self.logger.info(
            "deleting filter on IFID {0}, handle {1:X}".format(
                eth['name'], mark
            )
        )
        try:
            self.ipr.tc(
                RTM_DELTFILTER, 'fw', ifid, mark,
                parent=parent, protocol=ETH_P_IP, prio=PRIO
            )
        except NetlinkError as e:
            return TrafficControlRc(
                code=ReturnCode.NETLINK_FW_ERROR,
                message=str(e))
        except Exception as e:
            self.logger.exception('_unset_filter')
            exc_info = sys.exc_info()
            return TrafficControlRc(
                code=ReturnCode.UNKNOWN_FW_ERROR,
                message=str(exc_info))

        return TrafficControlRc(code=ReturnCode.OK)

    def _set_filter(self, mark, eth, shaping):
        """Given a mark, interface and shaping settings, create a TC filter.

        Args:
            mark: The mark based on which we create the filter.
            eth: The interface on which we create the filter.
            shaping: The shaping associated to this interface.

        Returns:
            A TrafficControlRc containing information on success/failure.
        """
        ifid = eth['id']
        idx = 0x10000 + mark
        parent = 0x10000
        self.logger.info(
            "create new FW filter on IFID {0}, classid {1},"
            " handle {2:X}, rate: {3}kbits".format(
                eth['name'], int_to_classid(idx), mark,
                shaping.rate
            )
        )
        try:
            extra_args = {}
            if not self.dont_drop_packets:
                extra_args.update({
                    'rate': "{}kbit".format(shaping.rate or 2**22 - 1),
                    'burst': self.burst_size,
                    'action': 'drop',
                })
            self.ipr.tc(RTM_NEWTFILTER, 'fw', ifid, mark,
                        parent=parent,
                        protocol=ETH_P_IP,
                        prio=PRIO,
                        classid=idx,
                        **extra_args
                        )
        except NetlinkError as e:
            return TrafficControlRc(
                code=ReturnCode.NETLINK_FW_ERROR,
                message=str(e))
        except Exception as e:
            self.logger.exception('_set_filter')
            exc_info = sys.exc_info()
            return TrafficControlRc(
                code=ReturnCode.UNKNOWN_FW_ERROR,
                message=str(exc_info))

        return TrafficControlRc(code=ReturnCode.OK)

    def _unset_iptables(self, mark, eth, ip, options=None):
        """Given a mark, interface, IP and options, clear iptables rules.

        Args:
            mark: The mark to delete.
            eth: The interface on which to delete the mark.
            ip: The IP address to shape.
            options: An array of iptables options for more specific filtering.

        Returns:
            A TrafficControlRc containing information on success/failure.
        """
        if options is None or len(options) == 0:
            options = ['']
        for opt in options:
            cmd = "{0} -t mangle -D FORWARD {1} {2} -i {3} {option} " \
                "-j MARK --set-mark {4}".format(
                    self.iptables, "-s"
                    if eth['name'] == self.lan['name'] else "-d",
                    ip, eth['name'], mark, option=opt)
            self.run_cmd(cmd)

    def _set_iptables(self, mark, eth, ip, options=None):
        """Given a mark, interface, IP and options, create iptables rules.

        Those rules will mark packets which will be filtered by TC filter and
        put in the right shaping bucket.

        Args:
            mark: The mark to delete.
            eth: The interface on which to delete the mark.
            ip: The IP address to shape.
            options: An array of iptables options for more specific filtering.

        Returns:
            A TrafficControlRc containing information on success/failure.
        """
        if options is None or len(options) == 0:
            options = ['']
        for opt in options:
            cmd = "{0} -t mangle -A FORWARD {1} {2} -i {3} {option} " \
                "-j MARK --set-mark {4}".format(
                    self.iptables, "-s"
                    if eth['name'] == self.lan['name'] else "-d",
                    ip, eth['name'], mark, option=opt)
            self.run_cmd(cmd)

    def _shape_interface(self, mark, eth, ip, shaping):
        """Shape the traffic for a given interface.

        Shape the traffic for a given IP on a given interface, given the mark
        and the shaping settings.
        There is a few steps to shape the traffic of an IP:
        1. Create an HTB class that limit the throughput.
        2. Create a NetEm QDisc that adds corruption, loss, reordering, loss
            and delay.
        3. Create the TC filter that will bucket packets with a given mark in
            the right HTB class.
        4. Set an iptables rule that mark packets going to/coming from IP

        Args:
            mark: The mark to set on IP packets.
            eth: The network interface.
            ip: The IP to shape traffic for.
            shaping: The shaping setting to set.

        Returns:
            A TrafficControlRc containing information on success/failure.
        """
        self.logger.info(
            "Shaping ip {0} on interface {1}".format(ip, eth['name']))
        # HTB class
        tcrc = self._set_htb_class(mark, eth, shaping)
        if tcrc.code != ReturnCode.OK:
            self.logger.error(
                "adding HTB class on IFID {0}, mark {1}, err: {2}".format(
                    eth['name'], mark, tcrc.message))
            return tcrc
        # NetemQdisc
        tcrc = self._set_netem_qdisc(mark, eth, shaping)
        if tcrc.code != ReturnCode.OK:
            self.logger.error(
                "adding NetEm qdisc on IFID {0}, mark {1}, err: {2}".format(
                    eth['name'], mark, tcrc.message))
            # delete class
            self._unset_htb_class(mark, eth)
            return tcrc
        # filter
        tcrc = self._set_filter(mark, eth, shaping)
        if tcrc.code != ReturnCode.OK:
            self.logger.error(
                "adding filter FW on IFID {0}, mark {1}, err: {2}".format(
                    eth['name'], mark, tcrc.message))
            # delete class
            self._unset_htb_class(mark, eth)
            return tcrc
        # iptables
        self._set_iptables(mark, eth, ip, shaping.iptables_options)

        return TrafficControlRc(code=ReturnCode.OK)

    def _unshape_interface(self, mark, eth, ip, settings):
        """Unshape the traffic for a given interface.

        Unshape the traffic for a given IP on a given interface, given the mark
        and the shaping settings.
        There is a few steps to unshape the traffic of an IP:
        1. Remove the iptables rule.
        2. Remove the TC filter.
        3. Remove the HTB class.

        Args:
            mark: The mark to set on IP packets.
            eth: The network interface.
            ip: The IP to shape traffic for.
            shaping: The shaping setting to set.

        Returns:
            A TrafficControlRc containing information on success/failure.
        """

        self.logger.info(
            "Unshaping ip {0} on interface {1}".format(ip, eth['name']))
        # iptables
        self._unset_iptables(mark, eth, ip, settings.iptables_options)
        # filter
        tcrc = self._unset_filter(mark, eth)
        if tcrc.code != ReturnCode.OK:
            self.logger.error(
                "deleting FW filter on IFID {0}, mark {1}, err: {2}".format(
                    eth['name'], mark, tcrc.message)
            )
            return tcrc
        # HTB class
        tcrc = self._unset_htb_class(mark, eth)
        if tcrc.code != ReturnCode.OK:
            self.logger.error(
                "deleting HTB class on IFID {0}, mark {1}, err: {2}".format(
                    eth['name'], mark, tcrc.message)
            )
            return tcrc

        return TrafficControlRc(code=ReturnCode.OK)
