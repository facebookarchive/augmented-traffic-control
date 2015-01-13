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
# Definition:: install_packages
#

define :install_packages, :packages => [] do
  params[:packages].each do |pkg|
    package pkg do
      action :install
    end
  end
end

define :install_virtualenv_packages, :packages => [], :virtualenv => nil do
  params[:packages].each do |k, v|
    python_pip k do
      version v[:version] if v.key?(:version)
      action v[:action] if v.key?(:action)
      options v[:options] if v.key?(:options)
      virtualenv params[:virtualenv] unless  v.fetch(:global, false)
    end
  end
end
