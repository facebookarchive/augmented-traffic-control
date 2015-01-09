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
# Recipe:: atcui
#
include_recipe 'atc::_common_system'
include_recipe 'atc::_virtualenv'

atc_user=node['atc']['user']
atc_group=node['atc']['group']

# Set python environment.
install_virtualenv_packages 'atcui_packages' do
    packages node['atc']['venv']['atcui']['packages']
    virtualenv node['atc']['venv']['path']
end

django_root = File.dirname(node['atc']['atcui']['base_dir'])
django_project = File.basename(node['atc']['atcui']['base_dir'])

directory django_root do
  owner atc_user
  group atc_group
  mode 00755
  recursive true
end

directory '/var/log/atc' do
  owner atc_user
  group atc_group
  mode 00750
end

execute 'install django' do
  command "#{File.join(node['atc']['venv']['path'], 'bin', 'django-admin')} startproject #{django_project} ."
  cwd django_root
  user atc_user
  group atc_group
  not_if { ::File.exists?(File.join(django_root, 'manage.py')) }
end

%w{urls settings}.each do |file|
  template File.join(node['atc']['atcui']['base_dir'], "#{file}.py") do
    source "django/#{file}.py.erb"
    mode 0644
    owner atc_user
    group atc_group
    notifies :restart, 'service[atcui]', :delayed
  end
end

template node['atc']['atcui']['config_file'] do
  source 'config/atcui.erb'
  mode 0755
  owner atc_user
  group atc_group
  notifies :restart, 'service[atcui]', :delayed
end

cookbook_file '/etc/init.d/atcui' do
  source "init.d/atcui.#{node['platform_family']}"
  mode 0755
  owner 'root'
  group 'root'
  notifies :restart, 'service[atcui]', :delayed
end

service 'atcui' do
  supports :restart => true
  action [:enable, :start]
end

template '/usr/local/bin/atcui-setup' do
  source 'atcui-setup.erb'
  mode 0755
  owner atc_user
  group atc_group
end

if node.vagrant?
  # When running under vagrant, atcui depends on the mounts and will not start
  # unless those are up. The mount is happening after the system is up.
  # We can use udev to trigger starting/stopping atcui when amount/umount event
  # happens.
  template '/etc/udev/rules.d/50-vagrant-mount-atcui.rules' do
    source 'mount-udev.rules.erb'
    variables({
      :service => 'atcui'
    })
  end
end


