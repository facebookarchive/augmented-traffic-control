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
      client: new AtcRestClient(this.props.primary, this.props.secondary, this.props.endpoint),
      profiles: null,
      potential: null,
      current: null,
      changed: false,
    };
  },

  componentDidMount: function() {
    this.updateState();
    this.update_interval = setInterval(this.updateState, 3000);
  },

  componentWillUnmount: function() {
    clearInterval(this.update_interval);
  },

  updateState: function() {
    this.fetchProfiles();
    this.fetchShaping();
  },

  createNewProfile: function(name) {
    // FIXME SETTINGS
    this.state.client.createProfile(
      {name:name, shaping:this.state.potential.shaping}, function(rc) {
        if (rc.status == 200) {
          this.fetchProfiles();
        }
      }.bind(this)
    );
  },

  fetchProfiles: function() {
    this.state.client.getProfiles(function (rc) {
      if (rc.status == 200) {
        this.setState({
          profiles: rc.json.profiles,
        });
      }
    }.bind(this));
  },

  selectProfile: function(shaping) {
    this.setState({
      potential: {shaping: shaping},
      changed: true,
    });
  },

  fetchShaping: function() {
    this.state.client.getShaping(function(rc) {
      var current = null;
      if (rc.status != 200) {
        current = null;
      } else {
        current = rc.json;
      }
      this.setState({current: current});
      if (this.state.changed) {
        // Don't overwrite the user-provided info in potential
        return;
      }
      if (rc.status != 200 || rc.json.shaping == null) {
        this.setState({potential: {shaping: defaultShaping()}});
      } else {
        this.setState({potential: rc.json});
      }
    }.bind(this));
  },

  performShaping: function() {
    this.state.client.shape(this.state.potential, function(rc) {
      if (rc.status == 200) {
        this.setState({
          current: rc.json.shaping,
          potential: rc.json.shaping,
          changed: false,
        });
      }
    }.bind(this));
  },

  clearShaping: function() {
    this.state.client.unshape(function(rc) {
      if (rc.status == 204) {
        // Notify unshaped successfully
        this.setState({
          current: null,
        });
      }
    }.bind(this));
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
        <CollapsePanel title="Profiles">
          <ProfilePanel profiles={this.state.profiles} onSave={this.createNewProfile} onSelect={this.selectProfile} />
        </CollapsePanel>
        <CollapsePanel title="Shaping">
          <ShapingPanel current={this.state.current} potential={this.state.potential} shapingDisabled={!this.state.changed} onPerformShaping={this.performShaping} onClearShaping={this.clearShaping} onSetPotential={this.selectProfile} />
        </CollapsePanel>
      </div>
    );
  },
});
