/**
 * Copyright (c) 2014, Facebook, Inc.
 * All rights reserved.
 *
 *  This source code is licensed under the BSD-style license found in the
 *  LICENSE file in the root directory of this source tree. An additional grant
 *  of patent rights can be found in the PATENTS file in the same directory.
 */

var ServerInfoPanel = React.createClass({
  getInitialState: function() {
    return {
      info: null,
    };
  },

  componentDidMount: function() {
    this.updateInfo();
    this.update_interval = setInterval(this.updateInfo, 1000);
  },

  componentWillUnmount: function() {
    clearInterval(this.update_interval);
  },

  updateInfo: function() {
    this.props.client.getServerInfo(function(rc) {
      if (rc.status == 200) {
        this.setState(function(state, props) {
          return {
            info: rc.json,
          }
        });
      } else {
        this.setState(function(state, props) {
          return {
            info: null,
          }
        });
      }
    }.bind(this));
  },

  render: function() {
    if (this.state.info == null) {
      return (
        <div>
          <i>ATC is not running.</i>
        </div>
      );
    } else {
      return (
        <div>
          <div>
            API Version: <code>{this.state.info.atc_api.version}</code>
          </div>
          <div>
            Daemon Version: <code>{this.state.info.atc_daemon.version}</code>
          </div>
          <div>
            Platform Type: <code>{this.state.info.atc_daemon.platform}</code>
          </div>
        </div>
      );
    }
  },
});
