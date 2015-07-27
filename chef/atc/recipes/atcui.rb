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

atcui_user = node['atc']['atcui']['user']
actui_group = node['atc']['atcui']['group']

# Set python environment.
atc_install_virtualenv_packages 'atcui_packages' do
  packages node['atc']['venv']['atcui']['packages']
  virtualenv node['atc']['venv']['path']
end

django_root = File.dirname(node['atc']['atcui']['base_dir'])
django_project = File.basename(node['atc']['atcui']['base_dir'])

directory django_root do
  owner atcui_user
  group actui_group
  mode 00755
  recursive true
end

directory '/var/log/atc' do
  owner atcui_user
  group actui_group
  mode 00750
end

execute 'install django' do
  command "#{File.join(node['atc']['venv']['path'], 'bin', 'django-admin')} " \
    "startproject #{django_project} ."
  cwd django_root
  user atcui_user
  group actui_group
  not_if { ::File.exist?(File.join(django_root, 'manage.py')) }
end

%w(urls settings).each do |file|
  template File.join(node['atc']['atcui']['base_dir'], "#{file}.py") do
    source "django/#{file}.py.erb"
    mode 0644
    owner atcui_user
    group actui_group
    notifies :restart, 'service[atcui]', :delayed
  end
end

template node['atc']['atcui']['config_file'] do
  source 'config/atcui.erb'
  mode 0755
  owner atcui_user
  group actui_group
  notifies :restart, 'service[atcui]', :delayed
end

if Chef::Platform::ServiceHelpers.service_resource_providers.include? :upstart
  template '/etc/init/atcui.conf' do
    source 'upstart/atcui.conf.erb'
    mode 0644
    owner 'root'
    group 'root'
    notifies :restart, 'service[atcui]', :delayed
  end
elsif Chef::Platform::ServiceHelpers.service_resource_providers.include? \
  :systemd
  log 'systemd not currently supported.' do
    level :warn
  end
else
  log 'Unsupported init system: ' +
    Chef::Platform::ServiceHelpers.service_resource_providers.to_s do
    level :warn
  end
end

service 'atcui' do
  provider Chef::Provider::Service::Upstart
  supports :restart => true
  action [:enable, :start]
end

atcui_setup_file = '/usr/local/bin/atcui-setup'
template atcui_setup_file do
  source 'atcui-setup.erb'
  mode 0755
  owner atcui_user
  group actui_group
end

# Setup atcui if it is the first time it is installed
atcui_configured_file = '/.atcui_configured'
unless File.exist?(atcui_configured_file)
  execute 'atcui setup' do
    command atcui_setup_file
    user atcui_user
    group actui_group
  end
  file atcui_configured_file do
    action :touch
  end
end
# When running under vagrant, atcui depends on the mounts and will not start
# unless those are up. The mount is happening after the system is up.
# We can use udev to trigger starting/stopping atcui when amount/umount event
# happens.
template '/etc/udev/rules.d/50-vagrant-mount-atcui.rules' do
  only_if { node.vagrant? }
  source 'mount-udev.rules.erb'
  variables(
    :service => 'atcui'
  )
end
