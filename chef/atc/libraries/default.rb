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
  class Node
    def virtualized?
      if self.key?('virtualization') and \
        self.virtualization.key?('system')
        return true
      end
      return false
    end

    def vagrant?
      return self.virtualized? && self.etc.passwd.key?('vagrant')
    end

    def get_default_user
      if self.vagrant?
        return 'vagrant'
      end
      return 'root'
    end

    def get_repo_basedir
      return node['atc']['src_dir']
    end

    def selinux?
        if File.exists?('/selinux/enforce')
            return File.read('/selinux/enforce') == '1'
        end
        return false
    end
  end
end
