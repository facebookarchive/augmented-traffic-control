#
#  Copyright (c) 2014, Facebook, Inc.
#  All rights reserved.
#
#  This source code is licensed under the BSD-style license found in the
#  LICENSE file in the root directory of this source tree. An additional grant
#  of patent rights can be found in the PATENTS file in the same directory.
#
#
name             'atc'
maintainer       'Facebook, Inc.'
maintainer_email ''
license          'BSD'
description      'Installs/Configures atc'
long_description IO.read(File.join(File.dirname(__FILE__), 'README.md'))
version          '0.1.0'

depends 'apt'
depends 'dhcp', '~> 2.2.2'
depends 'python'
depends 'simple_iptables', '~> 0.6.4'
depends 'sysctl', '~> 0.6.0'
depends 'yum-epel'
