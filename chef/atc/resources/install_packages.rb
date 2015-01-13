#
#  Copyright (c) 2014, Facebook, Inc.
#  All rights reserved.
#
#  This source code is licensed under the BSD-style license found in the
#  LICENSE file in the root directory of this source tree. An additional grant
#  of patent rights can be found in the PATENTS file in the same directory.
#
#
#
# Cookbook Name:: atc
# Resources:: install_packages
#

actions :install
attribute :packages, :kind_of => Array, :default => []
default_action :install
