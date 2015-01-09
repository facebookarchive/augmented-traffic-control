#
#  Copyright (c) 2014, Facebook, Inc.
#  All rights reserved.
#
#  This source code is licensed under the BSD-style license found in the
#  LICENSE file in the root directory of this source tree. An additional grant
#  of patent rights can be found in the PATENTS file in the same directory.
#
#
# Cookbook Name:: atc
# Recipe:: _common_system
#

case node['platform_family']
when 'rhel'
  include_recipe 'yum-epel'
end

group node['atc']['group'] do
    system
end

user node['atc']['user'] do
	system
    gid node['atc']['group']
	shell "/sbin/nologin"
end

case node['platform_family']
when 'rhel'
	execute 'yum makecache'
when 'debian'
	execute 'apt-get update'
else
	log "Not updating package cache." do
		level :warn
	end
end

install_packages "p" do
  packages node['atc']['packages']
end

case node['platform_family']
when "rhel"
    service "ntpd" do
        action [:enable, :start]
    end
end
