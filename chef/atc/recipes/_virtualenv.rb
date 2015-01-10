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
# Recipe:: _virtualenv
#
atcui_user=node['atc']['atcui']['user']
atcui_group=node['atc']['atcui']['group']

directory node['atc']['base_dir'] do
    owner atcui_user
    group atcui_group
    mode 00755
    action :create
    recursive true
end

# FIXME
# setuptools fail to install a package in virtual mode as it looks for
# Makefile in /usr/atc...
# Original error:
# STDERR: /usr/local/atc/venv/local/lib/python2.7/site-packages/pip/pep425tags.py:62: RuntimeWarning: invalid Python installation: unable to open /usr/atc/venv/lib/python2.7/config/Makefile (No such file or directory)
#  warnings.warn("{0}".format(e), RuntimeWarning)
link node['atc']['base_dir'].gsub('/usr/local', '/usr') do
    to node['atc']['base_dir']
end
 
python_virtualenv node['atc']['venv']['path'] do
    interpreter node['atc']['venv']['interpreter']
    owner atcui_user
    group atcui_group
    action :create
end
