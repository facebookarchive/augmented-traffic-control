#
#  Copyright (c) 2014, Facebook, Inc.
#  All rights reserved.
#
#  This source code is licensed under the BSD-style license found in the
#  LICENSE file in the root directory of this source tree. An additional grant
#  of patent rights can be found in the PATENTS file in the same directory.
#
#
# Monkey patch to allow installing unbuntu-desktop without timeouting until it is fixed upstream
# see https://github.com/test-kitchen/test-kitchen/issues/380#issuecomment-41359083
class ::Chef::Provider::Package::Apt
  def run_noninteractive(command)
    # There are some mighty big packages in this recipe, and 600s is just not enough!
    shell_out!(command, :env => { "DEBIAN_FRONTEND" => "noninteractive", "LC_ALL" => nil }, :log_level => :info, :timeout => 216000)
  end
end

