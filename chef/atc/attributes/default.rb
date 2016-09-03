#
#  Copyright (c) 2014, Facebook, Inc.
#  All rights reserved.
#
#  This source code is licensed under the BSD-style license found in the
#  LICENSE file in the root directory of this source tree. An additional grant
#  of patent rights can be found in the PATENTS file in the same directory.
#
#
packages = %w(
  git
)
base_dir = '/usr/local/atc'
src_dir = '/usr/local/src/atc'

default['atc']['src_dir'] = src_dir
default['atc']['packages'] = packages
default['atc']['base_dir'] = base_dir
default['atc']['venv']['interpreter'] = 'python2.7'
default['atc']['venv']['path'] = File.join(base_dir, 'venv')
default['atc']['venv']['atcd']['packages'] = {
  "file://#{File.join(src_dir, 'atc/atc_thrift/')}" =>
    { :action => :install, :options => '-e' },
  "file://#{File.join(src_dir, 'atc/atcd/')}" =>
    { :action => :install, :options => '-e' }
}

# 'pyroute2' => {:action => :upgrade, :version => "0.1.12"},
default['atc']['venv']['atcui']['packages'] = {
  'django' => { :version => '1.10' },
  'gunicorn' => {},
  "file://#{File.join(src_dir, 'atc/atc_thrift/')}" =>
    { :action => :install, :options => '-e' },
  "file://#{File.join(src_dir, 'atc/django-atc-api')}" =>
    { :action => :install, :options => '-e' },
  "file://#{File.join(src_dir, 'atc/django-atc-demo-ui')}" =>
    { :action => :install, :options => '-e' },
  "file://#{File.join(src_dir, 'atc/django-atc-profile-storage')}" =>
    { :action => :install, :options => '-e' },
  'mock' => {}
}

default['atc']['atcd']['listen']['address'] = '0.0.0.0'
default['atc']['atcd']['listen']['port'] = '9090'
default['atc']['atcd']['sqlite'] = '/var/lib/atc/atcd.db'
default['atc']['atcd']['interface']['wan'] = 'eth0'
default['atc']['atcd']['interface']['lan'] = 'eth1'
default['atc']['atcd']['mode'] = 'secure'

default['atc']['atcui']['workers'] = 2
default['atc']['atcui']['listen']['address'] = '0.0.0.0'
default['atc']['atcui']['listen']['port'] = '8000'
default['atc']['atcui']['base_dir'] = '/var/django/atc_ui'
default['atc']['atcui']['user'] = 'atc'
default['atc']['atcui']['group'] = 'atc'

case node['platform_family']
when 'rhel'
  default['atc']['atcui']['config_file'] = '/etc/sysconfig/atcui'
  default['atc']['atcd']['config_file'] = '/etc/sysconfig/atcd'
when 'debian'
  default['atc']['atcui']['config_file'] = '/etc/default/atcui'
  default['atc']['atcd']['config_file'] = '/etc/default/atcd'
end

# django app settings
default['atc']['atc_api']['default_timeout'] = 24 * 60 * 60

if platform?('centos') && node['platform_version'].to_i < 7
  # setuid not supported by the upstart version centos6 uses.
  default['atc']['atcui']['user'] = 'root'
  default['atc']['atcui']['group'] = 'root'
end
