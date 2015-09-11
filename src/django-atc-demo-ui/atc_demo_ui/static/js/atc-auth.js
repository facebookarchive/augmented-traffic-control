/** @jsx React.DOM */
/**
 * Copyright (c) 2014, Facebook, Inc.
 * All rights reserved.
 *
 *  This source code is licensed under the BSD-style license found in the
 *  LICENSE file in the root directory of this source tree. An additional grant
 *  of patent rights can be found in the PATENTS file in the same directory.
 */


var TokenFrame = React.createClass({
  getInitialState: function() {
    this.oos_notified = false;
    return {
      token: null,
      valid_until: null,
    };
  },

  componentDidMount: function() {
    this.getToken();
    this.interval = setInterval(this.getToken, 3000);
  },

  componentWillUnmount: function() {
    if (this.interval != null) {
      clearInterval(this.interval);
    }
  },

  getToken: function() {
    this.props.client.getToken(function (result) {
      if (result.status >= 200 && result.status < 300) {
        valid_until = new Date(result.json.valid_until*1000).toLocaleTimeString();
        if (result.json.valid_until - Math.floor(new Date().getTime() / 1000) < 0) {
          if (!this.oos_notified) {
            this.props.notify("warn", "The time on the ATC server is out of sync.");
            this.oos_notified = true;
          }
        }
        this.setState({
          token: result.json,
        });
      } else {
        this.props.notify("error", "Could not fetch current token: " + result.json);
        this.setState({
          token: null,
        });
      }
    }.bind(this));
  },

  render: function() {
    if (this.state.token == null) {
      return null;
    }

    return (
      <div className="col-md-6">
        <div>
          <h4>This Machine&#39;s Token: <b>{this.state.token.token}</b></h4>
          <b>Valid Until:</b> {valid_until}
          <h4>This Machine&#39; Address: {this.state.token.address}</h4>
        </div>
      </div>
    );
  },
});

var AuthFrame = React.createClass({
  getInitialState: function() {
    return {
      auth: null,
      token: null,
      address: null,
    };
  },

  componentDidMount: function() {
    this.getAuthInfo();
  },

  updateToken: function(event) {
    this.setState({token: event.target.value});
  },

  updateAddress: function(event) {
    this.setState({address: event.target.value});
  },

  getAuthInfo: function() {
    this.props.client.getAuthInfo(function (result) {
      if (result.status >= 200 && result.status < 300) {
        this.setState({
          auth: result.json,
          address: result.json.address,
        });
      } else {
        this.props.notify("error", "Could not fetch auth info: " + result.json);
        this.setState({
          auth: null,
          address: null,
        });
      }
    }.bind(this));
  },

  updateAuth: function() {
    var failed = false;
    if (this.state.address == null || this.state.address == "") {
      this.props.notify("error", "You must enter an address");
      failed = true;
    }
    if (this.state.token == null || this.state.token == "") {
      this.props.notify("error", "You must enter a token");
      failed = true;
    }
    if (failed) {
      return;
    }
    this.props.client.updateAuthInfo(this.state.address, {token: Number(this.state.token)}, function(result) {
      if (result.status >= 200 && result.status < 300) {
        console.log("Authorizing:", result.json);
        this.props.notify("success", "You can now shape " + result.json.controlled_ip);
      } else {
        this.props.notify("error", "Could not update auth info: ", result.json);
      }
    }.bind(this));
  },

  render: function() {
    if (this.state.auth == null) {
      return null;
    }

    var controlled_ips = null;
    if (this.state.auth.controlled_ips.length > 0) {
      controlled_ips = this.state.auth.controlled_ips.map(function (addr) {
        return (
          <li><pre><code>{addr}</code></pre></li>
        );
      });
      controlled_ips = (
        <ul>{controlled_ips}</ul>
      );
    } else {
      controlled_ips = (
        <i>No Controlled Machines</i>
      );
    }

    return (
      <div className="col-md-6">
        <div>
          <h4>Machines You Can Shape:</h4>
          {controlled_ips}
          <p>
          <b>Note:</b> A machine is always allowed to shape itself.
          </p>

          <h4>Authorize a New Machine:</h4>
          <label className="control-label">Address:</label>
          <input type="text" className="form-control" placeholder="127.0.0.1" onChange={this.updateAddress}/>
          <label className="control-label">Token:</label>
          <input type="number" className="form-control" placeholder="12345" onChange={this.updateToken}/>
          <button className="btn btn-success" onClick={this.updateAuth}>Authorize</button>
        </div>
      </div>
    );
  },
});

var AuthPanel = React.createClass({
  render: function() {
    return (
      <div className="panel-group" id="accordion3" role="tablist" aria-multiselectable="false">
        <div className="panel panel-default">
          <div className="panel-heading" data-toggle="collapse" data-parent="#accordion3" href="#collapseAuth" aria-expanded="false" aria-controls="collapseAuth">
              <h4 className="panel-title">
                  Authentication
              </h4>
          </div>
          <div id="collapseAuth" className="panel-collapse collapse" role="tabpanel">
            <div className="panel-body">

              <div className="row">
                <AuthFrame client={this.props.client} notify={this.props.notify} />
                <TokenFrame client={this.props.client} notify={this.props.notify} />
              </div>
            </div>
          </div>
        </div>
      </div>
    );
  }
})