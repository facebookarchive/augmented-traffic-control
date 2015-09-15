/** @jsx React.DOM */
/**
 * Copyright (c) 2014, Facebook, Inc.
 * All rights reserved.
 *
 *  This source code is licensed under the BSD-style license found in the
 *  LICENSE file in the root directory of this source tree. An additional grant
 *  of patent rights can be found in the PATENTS file in the same directory.
 */

var NoGroup = React.createClass({
  getInitialState: function() {
    return {
      token: null,
      group_id: null,
    };
  },

  createGroupCB: function() {
    this.props.client.createGroup(function(rc) {
      if (rc.status == 200) {
        this.props.updateGroup();
      }
    }.bind(this));
  },

  updateTokenCB: function(event) {
    this.setState({token: event.target.value});
  },

  updateGroupIdCB: function(event) {
    this.setState({group_id: event.target.value});
  },

  joinGroupCB: function() {
    this.props.client.joinGroup(this.state.group_id, {token: this.state.token.toString()}, function(rc) {
      if (rc.status == 200) {
        this.props.updateGroup();
      }
    }.bind(this));
  },

  render: function() {
    return (
      <div>
        <div>
          <i>You are not in a group.</i>
        </div>

        <div className="row">
          <div className="col-md-6">
            <h2>Create a Group</h2>
            <button type="button" className="btn btn-info" onClick={this.createGroupCB}>Create New Group</button>
          </div>

          <div className="col-md-6">
            <h2>Join a Group</h2>
            <label className="control-label">Group ID:</label>
            <input type="number" className="form-control" placeholder="group id" onChange={this.updateGroupIdCB}/>
            <label className="control-label">Token:</label>
            <input type="number" className="form-control" placeholder="token" onChange={this.updateTokenCB}/>
            <button className="btn btn-success" onClick={this.joinGroupCB}>Join Group</button>
          </div>
        </div>
      </div>
    );
  },
});

var InGroup = React.createClass({
  getInitialState: function() {
    return {
      token: null,
    };
  },

  componentDidMount: function() {
    this.updateToken();
    this.update_interval = setInterval(this.updateToken, 10000); // 10s
  },

  componentWillUnmount: function() {
    if (this.update_interval != null) {
      clearInterval(this.update_interval);
    }
  },

  updateToken: function() {
    this.props.client.getToken(this.props.group.id, function(rc) {
      if (rc.status == 200) {
        this.setState(function(state, props) {
          return {
            token: rc.json,
          };
        });
      }
    }.bind(this));
  },

  leaveGroupCB: function() {
    this.props.client.leaveGroup(this.state.token.id, this.state.token, function(rc) {
      if (rc.status == 200) {
        this.props.updateGroup();
      }
    }.bind(this))
  },

  render: function() {
    var memberNodes = this.props.group.members.map(function(item, idx, arr) {
      return (
        <li><code>{item}</code></li>
      );
    });
    var token = null;
    if (this.state.token != null) {
      token = (
        <span>Token: <code>{this.state.token.token}</code><br/></span>
      );
    }
    return (
      <div>
        Group ID: {this.props.group.id}<br/>
        Members:<br/>
        <ul>
        {memberNodes}
        </ul>
        {token}
        <button type="button" className="btn btn-warning" onClick={this.leaveGroupCB}>Leave Group</button>
      </div>
    );
  },
});

var GroupPanel = React.createClass({
  getInitialState: function() {
    return {
      group: null,
    };
  },

  componentDidMount: function() {
    this.update_interval = setInterval(this.updateGroup, 1000);
  },

  componentWillUnmount: function() {
    clearInterval(this.update_interval);
  },

  updateGroup: function() {
    // Get group from API
    this.props.client.getGroup(function(rc) {
      if (rc.status == 200) {
        this.setState(function(state, props) {
          return {
            group: rc.json,
          };
        });
      } else if (rc.status == 404) {
        this.setState(function(state, props) {
          return {
            group: null,
          };
        });
      }
    }.bind(this));
  },

  render: function() {
    if (this.state.group != null) {
      return (
        <InGroup events={this.props.events} client={this.props.client} group={this.state.group} updateGroup={this.updateGroup} />
      );
    } else {
      return (
        <NoGroup events={this.props.events} client={this.props.client} updateGroup={this.updateGroup} />
      );
    }
  },
});
