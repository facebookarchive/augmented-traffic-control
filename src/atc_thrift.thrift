#
#  Copyright (c) 2015, Facebook, Inc.
#  All rights reserved.
#
#  This source code is licensed under the BSD-style license found in the
#  LICENSE file in the root directory of this source tree. An additional grant
#  of patent rights can be found in the PATENTS file in the same directory.
#

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

struct Setting {
  1: Shaping up,
  2: Shaping down,
}

enum PlatformType {
    LINUX = 0
}

struct AtcdInfo {
    1: PlatformType platform,
    2: string version,
}

struct ShapingGroup {
    1: i64 id,
    2: list<string> members,
    3: optional Setting shaping,
}

service Atcd {
    AtcdInfo get_atcd_info(),

    ShapingGroup create_group(1: string member),

    ShapingGroup get_group(1: i64 id),
    ShapingGroup get_group_with(1: string member),

    string get_group_token(1: i64 id),

    void leave_group(1: i64 id, 2: string to_remove, 3: string token),
    void join_group(1: i64 id, 2: string to_add, 3: string token),

    Setting shape_group(1: i64 id, 2: Setting settings, 3: string token),
    void unshape_group(1: i64 id, 2: string token),
}
