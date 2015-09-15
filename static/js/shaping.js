/** @jsx React.DOM */
/**
 * Copyright (c) 2014, Facebook, Inc.
 * All rights reserved.
 *
 *  This source code is licensed under the BSD-style license found in the
 *  LICENSE file in the root directory of this source tree. An additional grant
 *  of patent rights can be found in the PATENTS file in the same directory.
 */

var ShapingPanel = React.createClass({
  getInitialState: function() {
    return {
      changed: false,
      current: null,
      potential: null,
    };
  },

  componentDidMount: function() {
    this.updateShaping();
    this.update_interval = setInterval(this.updateShaping, 1000);
  },

  componentWillUnmount: function() {
    clearInterval(this.update_interval);
  },

  updateShaping: function() {
    this.props.client.getShaping(function(rc) {
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
        this.setState({potential: {shaping: defaultSettings()}, changed: true});
      } else {
        this.setState({potential: rc.json});
      }
    }.bind(this));
  },

  performShaping: function() {
    this.props.client.shape(this.state.potential, function(rc) {
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
    this.props.client.unshape(function(rc) {
      if (rc.status == 204) {
        // Notify unshaped successfully
        this.setState({
          current: null,
        });
      }
    }.bind(this));
  },

  updatePotential: function(ev) {
    this.setState({potential: {shaping: JSON.parse(ev.target.value)}, changed: true});
  },

  render: function() {
    if (this.state.current != null && this.state.current.shaping != null) {
      var clear_enabled = true;
      var before_view = (
        <JSONView json={this.state.current.shaping} label="Current:" />
      );
    } else {
      var clear_enabled = false;
      var before_view = (
        <i>You are not being shaped.</i>
      );
    }
    if (this.state.potential != null) {
      var after_view = (
        <JSONEdit json={this.state.potential.shaping} onchange={this.updatePotential} />
      );
    } else {
      // Shouldn't happen (hopefully)
      var after_view = (
        <b>Something went wrong.</b>
      );
    }
    return (
      <div>
        Shaping goes here!
        <div>
          <div className="col-md-6">
            <div className="row">
              <div className="col-md-6">
                <h4>Current:</h4>
              </div>
              <div className="col-md-6">
                <button type="button" className="btn btn-danger" disabled={!clear_enabled} onClick={this.clearShaping}>
                  Clear Shaping
                </button>
              </div>
            </div>
            {before_view}
          </div>
          <div className="col-md-6">
            <div className="row">
              <div className="col-md-6">
                <h4>New:</h4>
              </div>
              <div className="col-md-6">
                <button type="button" className="btn btn-info" disabled={!this.state.changed} onClick={this.performShaping}>
                  Apply Shaping
                </button>
              </div>
            </div>
            {after_view}
          </div>
        </div>
      </div>
    );
  },
});
