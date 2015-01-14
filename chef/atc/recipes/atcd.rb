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
# Recipe:: atcd
#
include_recipe 'atc::_common_system'
include_recipe 'atc::_virtualenv'

# Set sysctl values.
node.default['sysctl']['params']['net']['ipv4']['ip_forward'] = 1
node.default['sysctl']['allow_sysctl_conf'] = true
node['network']['interfaces'].each do |name, eth|
  next if eth.encapsulation != 'Ethernet'
  node.default['sysctl']['params']['net']['ipv4']['conf']\
    [name]['arp_ignore'] = 1
  node.default['sysctl']['params']['net']['ipv4']['conf']\
    [name]['arp_announce'] = 2
end

# If we're in a sandbox, i.e. vagrant, we want to set up NAT
# But if we're not, we will let the user handle that themselves.
# We need NAT because we have just a bunch of VMs (atcclients) without a router
# so the atcd box need to handle the routing/NAT
if node.vagrant?

  # Try to guess the correct interfaces.
  lan_iface = ''
  wan_iface = node['network']['default_interface']
  node['network']['interfaces'].each do |name, eth|
    next if (eth.encapsulation != 'Ethernet') || (name == wan_iface)
    lan_iface = name
    break
  end

  if lan_iface == ''
    log 'You only have 1 network interface. '\
      'ATCD requires 2 network interfaces. '\
      'Chef will likely fail.'
    lan_iface = 'eth1'
  end

  log "Using interfaces lan=>#{lan_iface}, wan=>#{wan_iface}"

  # NAT
  include_recipe 'simple_iptables'
  simple_iptables_rule 'nat' do
    table 'nat'
    direction 'POSTROUTING'
    jump 'MASQUERADE'
    rule "-o #{node['network']['default_interface']}"
  end

  # DHCP
  # By providing IP on the lan interface,
  # we can get Genymotion instances to route through atc
  require 'ipaddr'
  node.default['dhcp']['interfaces'] = [lan_iface]
  include_recipe 'dhcp::server'

  ipv4 = node['network']['interfaces'][lan_iface]['addresses'].select do |_, v|
    v.family == 'inet'
  end
  ip = ipv4.keys[0]
  ipaddr = IPAddr.new "#{ip}/#{ipv4[ip]['prefixlen']}"
  range_start = ipaddr | 100
  range_end = ipaddr | 200
  dhcp_subnet ipaddr.to_s do
    range "#{range_start} #{range_end}"
    broadcast ipv4[ip]['broadcast']
    netmask ipv4[ip]['netmask']
    routers [ip]
  end

  # When running under vagrant, atcd depends on the mounts and will not start
  # unless those are up. The mount is happening after the system is up.
  # We can use udev to trigger starting/stopping atcd when amount/umount event
  # happens.
  template '/etc/udev/rules.d/50-vagrant-mount-atcd.rules' do
    source 'mount-udev.rules.erb'
    variables(
      :service => 'atcd'
    )
  end
end

include_recipe 'sysctl::apply'

# Set python environment.
atc_install_virtualenv_packages 'atcd_packages' do
  packages node['atc']['venv']['atcd']['packages']
  virtualenv node['atc']['venv']['path']
end

# Ensure the Sqlite directory exists
directory File.dirname(node['atc']['atcd']['sqlite']) do
  owner 'root'
  group 'root'
  mode 00755
  recursive true
end

template node['atc']['atcd']['config_file'] do
  source 'config/atcd.erb'
  mode 0755
  owner 'root'
  group 'root'
  notifies :restart, 'service[atcd]', :delayed
end

cookbook_file '/etc/init.d/atcd' do
  source "init.d/atcd.#{node['platform_family']}"
  mode 0755
  owner 'root'
  group 'root'
  notifies :restart, 'service[atcd]', :delayed
end

service 'atcd' do
  supports :restart => true
  action [:enable, :start]
end
