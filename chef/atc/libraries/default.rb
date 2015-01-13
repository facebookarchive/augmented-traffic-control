#
#  Copyright (c) 2014, Facebook, Inc.
#  All rights reserved.
#
#  This source code is licensed under the BSD-style license found in the
#  LICENSE file in the root directory of this source tree. An additional grant
#  of patent rights can be found in the PATENTS file in the same directory.
#
#

class Chef
  # Some helper functions
  class Node
    def virtualized?
      if self.key?('virtualization') && \
         virtualization.key?('system')
        return true
      end
      false
    end

    def vagrant?
      virtualized? && etc.passwd.key?('vagrant')
    end

    def default_user
      return 'vagrant' if vagrant?
      'root'
    end

    def repo_basedir
      node['atc']['src_dir']
    end

    def selinux?
      if File.exist?('/selinux/enforce')
        return File.read('/selinux/enforce') == '1'
      end
      false
    end
  end
end
