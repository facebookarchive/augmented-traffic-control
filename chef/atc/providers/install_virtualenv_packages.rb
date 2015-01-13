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
# Provider:: install_virtualenv_packages
#

def handle_packages(packages, virtualenv)
  packages.each do |k, v|
    python_pip k do
      version v[:version] if v.key?(:version)
      action v[:action] if v.key?(:action)
      options v[:options] if v.key?(:options)
      virtualenv virtualenv unless  v.fetch(:global, false)
    end
  end
end

action :install do
  handle_packages(
    new_resource.packages,
    new_resource.virtualenv
  )
  new_resource.updated_by_last_action(true)
end
