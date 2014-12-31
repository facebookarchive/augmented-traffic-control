#
#  Copyright (c) 2014, Facebook, Inc.
#  All rights reserved.
#
#  This source code is licensed under the BSD-style license found in the
#  LICENSE file in the root directory of this source tree. An additional grant
#  of patent rights can be found in the PATENTS file in the same directory.
#
#
enum ReturnCode {
    OK = 0,
    INVALID_IP,
    INVALID_TIMEOUT,
    ID_EXHAUST,
    NETLINK_ERROR,
    UNKNOWN_ERROR,
    NETLINK_HTB_ERROR,
    UNKNOWN_HTB_ERROR,
    NETLINK_NETEM_ERROR,
    UNKNOWN_NETEM_ERROR,
    NETLINK_FW_ERROR,
    UNKNOWN_FW_ERROR,
    UNKNOWN_SESSION,
    UNKNOWN_IP,
    ACCESS_DENIED,
}

struct Delay {
  1: i32 delay,
  2: optional i32 jitter = 0,
  3: optional double correlation = 0,
}

struct Loss {
  1: double percentage,
  2: optional double correlation = 0,
}

struct Reorder {
  1: double percentage,
  2: i32 gap = 0,
  3: optional double correlation = 0,
}

struct Corruption {
  1: double percentage = 0,
  2: optional double correlation = 0,
}

struct Shaping {
  1: i32 rate,
  2: optional Delay delay = {"delay": 0},
  3: optional Loss loss = {"percentage": 0},
  4: optional Reorder reorder = {"percentage": 0},
  5: optional Corruption corruption = {"percentage": 0},
  6: optional list<string> iptables_options,
}

struct TrafficControlSetting {
  1: Shaping up,
  2: Shaping down,
}

struct TrafficControlledDevice {
  1: string controlledIP,
  2: optional string controllingIP,
}

struct RemoteControlInstance {
  1: TrafficControlledDevice device,
  2: optional i32 timeout,
}

struct TrafficControl {
  1: TrafficControlSetting settings,
  2: TrafficControlledDevice device,
  3: i32 timeout,
}

struct TrafficControlRc {
  1: ReturnCode code,
  2: optional string message,
}

exception TrafficControlException {
  1: ReturnCode code,
  2: optional string message,
}

exception PacketCaptureException {
  1: string message,
}

struct PacketCaptureFile {
  1: string name,
  2: string url,
  3: optional i32 bytes = 0,
}

struct PacketCapture {
  1: string ip,
  2: i32 start_time,
  3: PacketCaptureFile file,
  4: optional i32 pid = 0,
}

struct AccessToken {
  1: i32 token,
  2: i32 interval,
  3: i32 valid_until,
}

service Atcd {
  TrafficControlRc startShaping(1: TrafficControl tc)
    throws (1: TrafficControlException failure),
  TrafficControlRc stopShaping(1: TrafficControlledDevice device)
    throws (1: TrafficControlException failure),
  TrafficControl getCurrentShaping(1: TrafficControlledDevice device)
    throws (1: TrafficControlException failure),
  /* tell if whether of not an ip is shaped */
  bool isShaped(1: TrafficControlledDevice device)
    throws (1: TrafficControlException failure),
  PacketCapture startPacketCapture(1: TrafficControlledDevice device, 2: i32 timeout)
    throws (1: PacketCaptureException failure),
  PacketCapture stopPacketCapture(1: TrafficControlledDevice device)
    throws (1: PacketCaptureException failure),
  void stopAllPacketCaptures(),
  list<PacketCapture> listPacketCaptures(1: TrafficControlledDevice device)
    throws (1: TrafficControlException failure)
  list<PacketCapture> listRunningPacketCaptures(),
  /* returns the number of actively shaped devices */
  i32 getShapedDeviceCount(),
  /* Remote Control */
  AccessToken requestToken(1: string ip, 2: i32 duration),
  bool requestRemoteControl(1: TrafficControlledDevice device  2: AccessToken accessToken),
  list<RemoteControlInstance> getDevicesControlledBy(1: string ip),
  list<RemoteControlInstance> getDevicesControlling(1: string ip),
}
