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
# Provider:: install_packages
#
#
action :install do
  new_resource.packages.each do |pkg|
    package pkg do
      action :install
    end
  end
  new_resource.updated_by_last_action(true)
end
