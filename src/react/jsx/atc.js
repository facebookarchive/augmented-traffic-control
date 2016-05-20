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

function isIPv6(addr) {
    return addr.indexOf(':') >= 0;
}

var Atc = React.createClass({
  getInitialState: function() {
    this.update_intervals = {};
    return {
      client: null,
      profiles: null,
      potential: null,
      current: null,
      token: null,
      group: null,
      info: null,
      ipv4_ok: false,
      ipv6_ok: false,
      ipv4_shaped: false,
      ipv6_shaped: false,
    };
  },

  componentDidMount: function() {
    var client = new AtcRestClient(function() {
        this.setState({
          client: client,
          info: client.info,
        });
        this.updateToken();

        this.updateProfiles();
        this.update_intervals['profiles'] = setInterval(
          this.updateProfiles, 300000
        );
        this.updateInfo();
        this.update_intervals['info'] = setInterval(this.updateInfo, 300000);

        this.updateReacheability();
        this.update_intervals['reacheability'] = setInterval(
          this.updateReacheability, 6000
        );

        this.fetchGroup();
       this.update_intervals['group'] = setInterval(this.fetchGroup, 1000);

    }.bind(this));
  },

  componentWillUnmount: function() {
    for (update in this.update_intervals) {
      clearInterval(this.update_intervals['update']);
    }
  },

  /**
  * Update functions
  */
  updateInfo: function() {
    this.state.client.getServerInfo(function(rc) {
      if (rc.status == 200) {
        this.setState({info: rc.json});
      } else {
        this.setState({info: null});
      }
    }.bind(this));
  },

  updateReacheability: function() {
    console.log("updating reacheability");
    this.state.client.testIP(this.state.client.addresses['ipv4'], function(rc) {
        this.setState({ipv4_ok: rc});
    }.bind(this));
    this.state.client.testIP(this.state.client.addresses['ipv6'], function(rc) {
        this.setState({ipv6_ok: rc});
    }.bind(this));
  },

  updateProfiles: function() {
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

  /**
  * Profiles
  */
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
    });
  },

  /**
  * Groups
  */
  dualStackAccessible: function() {
    var ok = false;
    var addresses = this.state.client.addresses;
    if (this.state.client.dual_stack()) {
        if (addresses['secondary'].indexOf(':') >= 0 ? !this.state.ipv6_ok : !this.state.ipv4_ok) {
            console.warn('Dual-stacked but secondary IP is not accessible!');
        } else {
            ok = true;
        }
    } else {
        console.log('Not dual-stacked');
    }
    return ok;
  },

  createGroupCB: function() {
    var ok = true;
    var addresses = this.state.client.addresses;
    console.info('Creating group against ' + addresses['primary']);
    this.state.client.createGroup(function(rc) {
      if (rc.status == 200) {
        if (this.dualStackAccessible()){
          console.info('Dual stacked. joining group ' + rc.json.id + ' against ' + addresses['secondary']);
          this.state.client.joinGroupSecondary(
                rc.json.id,
                {token: rc.json.token.toString()},
                function(rc) { // eslint-disable-line no-unused-vars
            if (rc.status != 200) {
                console.error('Failed to join group on ' + addresses['secondary'] + ' endpoint with HTTP response ' + rc.status);
                ok = false;
            }
          });
        }
      } else {
        console.error('Failed to create group on ' + addresses['primary'] + 'endpoint with HTTP response ' + rc.status);
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
    var addresses = this.state.client.addresses;
    console.info('Leaving group ' + this.state.token.id + ' against ' + addresses['primary']);
    this.state.client.leaveGroup(this.state.token.id, this.state.token, function(rc) {
      if (rc.status == 200) {
        if (this.dualStackAccessible()) {
          console.info('Dual stacked. leaving group ' + rc.json.id + ' against ' + addresses['secondary']);
          this.state.client.leaveGroupSecondary(
                this.state.token.id,
                this.state.token,
                function(rc) { // eslint-disable-line no-unused-vars
            if (rc.status != 200) {
                console.error('Failed to leave group ' + this.state.token.id + ' against ' + addresses['secondary'] + ' endpoint with HTTP response ' + rc.status);
                ok = false;
            }
          }.bind(this));
        }
      } else {
        console.error('Failed to leave group ' + this.state.token.id + ' against ' + addresses['primary'] + ' endpoint with HTTP response ' + rc.status);
        ok = false;
      }
    }.bind(this));
    this.fetchGroup();
    return ok;
  },

  fetchGroup: function() {
    // Get group from API
    // Fixme... this only use primary addr
    var addresses = this.state.client.addresses;
    this.state.client.getGroupByAddr(addresses['primary'], function(xhr) {
      var shaped = false;
      if (xhr.status == 200) {
        this.setState({group: xhr.responseJSON});
        this.updateToken();
        shaped = true;
      } else if (xhr.status == 204) {
        this.setState({group: null, token: null});
      }
      if (isIPv6(addresses['primary'])) {
        this.setState({ipv6_shaped: shaped});
      } else {
        this.setState({ipv4_shaped: shaped});
      }
    }.bind(this));

    if (this.dualStackAccessible()) {
      this.state.client.getGroupByAddr(addresses['secondary'], function(xhr) {
        var shaped = false;
        if (xhr.status == 200) {
          shaped = true;
        }
        if (isIPv6(addresses['secondary'])) {
          this.setState({ipv6_shaped: shaped});
        } else {
          this.setState({ipv4_shaped: shaped});
        }
      }.bind(this));
    }

  },


  /**
  * Shaping
  */
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


  /**
  * Rendering
  */
  render: function () {
    return (
      <div>
        <CollapsePanel title="Shaping">
          <SimpleShapingPanel onToggleShaping={this.toggleShaping} profiles={this.state.profiles} onSelectProfile={this.selectProfile} shaped={this.state.group != null} potential_shaping={this.getPotentialShaping()} />
        </CollapsePanel>
        <CollapsePanel title="Server Info">
          <ServerInfoPanel info={this.state.info} ipv4_ok={this.state.ipv4_ok} ipv6_ok={this.state.ipv6_ok} ipv4_shaped={this.state.ipv4_shaped} ipv6_shaped={this.state.ipv6_shaped}/>
        </CollapsePanel>
      </div>
    );
  },
});

module.exports = Atc
