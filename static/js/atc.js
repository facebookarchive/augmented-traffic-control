/** @jsx React.DOM */
/**
 * Copyright (c) 2014, Facebook, Inc.
 * All rights reserved.
 *
 *  This source code is licensed under the BSD-style license found in the
 *  LICENSE file in the root directory of this source tree. An additional grant
 *  of patent rights can be found in the PATENTS file in the same directory.
 */

var ERROR_EXPIRY = 10000;

var atc_status = {
  OFFLINE: 0,
  ACTIVE: 1,
  INACTIVE: 2,
  OUTDATED: 3,
};


var NOTIFICATION_TYPES = {
  "error": "danger",
  "info": "info",
  "warn": "warning",
  "success": "success",
};


var NotificationPanel = React.createClass({
  render: function () {
    if (this.props.notifications.length == 0) {
      return (
        <i>No notifications.</i>
      );
    }
    return this.props.notifications.map(function(item, idx, arr) {
      var timeout = Math.floor((item.expire_at - new Date().getTime()) / 1000)
      var cls = "alert alert-" + (NOTIFICATION_TYPES[item.type] || item.type);
      return (
        <div className={cls} role="alert">
          <div className="row">
            <div className="col-md-11">{item.message}</div>
            <div className="col-md-1">{timeout}</div>
          </div>
        </div>
      );
    });
  },
});

var Atc = React.createClass({
  getInitialState: function() {
    return {
      client: new AtcRestClient(this.props.endpoint),
    };
  },

  render: function () {
    return (
      <div>
        <CollapsePanel title="Server Info">
          <ServerInfoPanel client={this.state.client} />
        </CollapsePanel>
        <CollapsePanel title="Group">
          <GroupPanel client={this.state.client} />
        </CollapsePanel>
        <CollapsePanel title="Shaping">
          <ShapingPanel client={this.state.client} />
        </CollapsePanel>
      </div>
    );
  },
});
