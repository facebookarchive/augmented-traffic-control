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
# Cookbook Name:: atcclient
# Recipe:: default
#
include_recipe 'apt'
package 'ubuntu-desktop'

# Force the window manager to start
service 'lightdm' do
  action [:enable, :start]
end
