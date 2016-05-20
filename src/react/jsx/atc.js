/**
 * Copyright (c) 2014, Facebook, Inc.
 * All rights reserved.
 *
 *  This source code is licensed under the BSD-style license found in the
 *  LICENSE file in the root directory of this source tree. An additional grant
 *  of patent rights can be found in the PATENTS file in the same directory.
 */

var React = require('react');
var AtcRestClient = require('./api');
var CollapsePanel = require('./utils').CollapsePanel;
var SimpleShapingPanel = require('./simple_shaping');
var ServerInfoPanel = require('./server');

var Atc = React.createClass({
  getInitialState: function() {
    return {
      client: new AtcRestClient(),
      profiles: null,
      potential: null,
      current: null,
      token: null,
      group: null,
    };
  },

  componentDidMount: function() {
    this.updateToken();
    this.updateState();
    this.update_state_interval = setInterval(this.updateState, 3000);
    this.fetch_group_interval = setInterval(this.fetchGroup, 1000);
  },

  componentWillUnmount: function() {
    if (this.update_state_interval) {
      clearInterval(this.update_state_interval);
    }
    if (this.fetch_group_interval != null) {
      clearInterval(this.fetch_group_interval);
    }
  },

  updateState: function() {
    this.fetchProfiles();
  },

  updateToken: function() {
    if ( this.state.group == null){
        return;
    }
    this.state.client.getToken(this.state.group.id, function(rc) {
      if (rc.status == 200) {
        this.setState({token: rc.json});
      }
    }.bind(this));
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
    console.log(shaping);
    this.setState({
      potential: {shaping: shaping},
    });
  },

  createGroupCB: function() {
    var ok = true;
    var addresses = this.state.client.addresses;
    console.info('Creating group against ' + addresses['primary']);
    this.state.client.createGroup(function(rc) {
      if (rc.status == 200) {
        if (this.state.client.dual_stack()) {
          console.info('Dual stacked. joining group ' + rc.json.id + ' against ' + addresses['secondary']);
          this.state.client.joinGroupSecondary(
                rc.json.id,
                {token: rc.json.token.toString()},
                function(rc) { // eslint-disable-line no-unused-vars
            if (rc.status != 200) {
                console.error('Failed to join group on' + addresses['secondary'] + ' endpoint with HTTP response ' + rc.status);
                ok = false;
            }
          });
        }
      } else {
        console.error('Failed to create group on' + addresses['primary'] + 'endpoint with HTTP response ' + rc.status);
        ok = false;
      }
    }.bind(this));
    if (ok) {
      this.fetchGroup();
    }
    return ok;
  },

  leaveGroupCB: function() {
    var ok = true;
    this.fetchGroup();
    var addresses = this.state.client.addresses;
    console.info('Leaving group ' + this.state.token.id + ' against ' + addresses['primary']);
    this.state.client.leaveGroup(this.state.token.id, this.state.token, function(rc) {
      if (rc.status == 200) {
        if (this.state.client.dual_stack()) {
          console.info('Dual stacked. leaving group ' + rc.json.id + ' against ' + addresses['secondary']);
          this.state.client.leaveGroupSecondary(
                this.state.token.id,
                this.state.token,
                function(rc) { // eslint-disable-line no-unused-vars
            if (rc.status != 200) {
                console.error('Failed to leave group ' + this.state.token.id + ' against ' + addresses['secondary'] + ' endpoint with HTTP response ' + rc.status);
                ok = false;
            }
          });
        }
      } else {
        console.error('Failed to leave group ' + this.state.token.id + ' against ' + addresses['primary'] + ' endpoint with HTTP response ' + rc.status);
        ok = false;
      }
    }.bind(this));
    return ok;
  },

  fetchGroup: function() {
    // Get group from API
    this.state.client.getGroup(function(rc) {
      if (rc.status == 200) {
        this.setState({group: rc.json});
        this.updateToken();
      } else if (rc.status == 204) {
        this.setState({group: null, token: null});
      }
    }.bind(this));
  },

  performShaping: function() {
    if (this.createGroupCB()) {
      this.state.client.shape(this.state.potential, function(rc) {
          if (rc.status == 200) {
              this.setState({
                current: rc.json.shaping,
                potential: {shaping: rc.json.shaping},
                changed: false,
              });
          }
      }.bind(this));
    }
  },

  clearShaping: function() {
    this.leaveGroupCB();
  },

  toggleShaping: function() {
    // If shaped, e.g we are in a group, unshaped
    if (this.state.group != null) {
      console.log('toggleShaping: unshaping');
      this.clearShaping();
    } else {
      if (this.getPotentialShaping() == null) {
        // Do nothing, should alert.
      } else {
        this.performShaping();
        console.log('toggleShaping: shaping');
      }
    }
  },

  getPotentialShaping: function() {
    if (this.state.potential == null || this.state.potential.shaping == null) {
      return null;
    }
    return this.state.potential.shaping;
  },

  render: function () {
    return (
      <div>
        <CollapsePanel title="Shaping">
          <SimpleShapingPanel onToggleShaping={this.toggleShaping} profiles={this.state.profiles} onSelectProfile={this.selectProfile} shaped={this.state.group != null} potential_shaping={this.getPotentialShaping()} />
        </CollapsePanel>
        <CollapsePanel title="Server Info">
          <ServerInfoPanel client={this.state.client} />
        </CollapsePanel>
      </div>
    );
  },
});

module.exports = Atc
