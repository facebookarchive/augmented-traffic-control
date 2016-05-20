/**
 * Copyright (c) 2014, Facebook, Inc.
 * All rights reserved.
 *
 *  This source code is licensed under the BSD-style license found in the
 *  LICENSE file in the root directory of this source tree. An additional grant
 *  of patent rights can be found in the PATENTS file in the same directory.
 */

var React = require('react');

var ServerInfoPanel = React.createClass({

  render: function() {
    function print_state(b) {
        return b ? 'OK' : 'NO';
    }

    if (this.props.info == null) {
      return (
        <div>
          <i>ATC is not running.</i>
        </div>
      );
    } else {
      return (
        <div>
          <div>
            IPv4: Connect <code>{print_state(this.props.ipv4_ok)}</code> Shaped: <code>{print_state(this.props.ipv4_shaped)}</code>
          </div>
          <div>
            IPv6: Connect <code>{print_state(this.props.ipv6_ok)}</code> Shaped: <code>{print_state(this.props.ipv6_shaped)}</code>
          </div>
            <div>&nbsp;</div>
          <div>
            API Version: <code>{this.props.info.atc_api.version}</code>
          </div>
          <div>
            Daemon Version: <code>{this.props.info.atc_daemon.version}</code>
          </div>
          <div>
            Platform Type: <code>{this.props.info.atc_daemon.platform}</code>
          </div>
        </div>
      );
    }
  },
});

module.exports = ServerInfoPanel
