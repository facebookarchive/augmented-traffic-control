/**
 * Copyright (c) 2014, Facebook, Inc.
 * All rights reserved.
 *
 *  This source code is licensed under the BSD-style license found in the
 *  LICENSE file in the root directory of this source tree. An additional grant
 *  of patent rights can be found in the PATENTS file in the same directory.
 */

var React = require('react');
var JSONEdit = require('./utils').JSONEdit;
var JSONView = require('./utils').JSONView;

var ShapingPanel = React.createClass({
  setPotential: function(ev) {
    this.props.onSetPotential(JSON.parse(ev.target.value));
  },

  render: function() {
    if (this.props.current != null && this.props.current.shaping != null) {
      var clear_enabled = true;
      var before_view = (
        <JSONView json={this.props.current.shaping} label="Current:" />
      );
    } else {
      var clear_enabled = false;
      var before_view = (
        <i>You are not being shaped.</i>
      );
    }
    if (this.props.potential != null) {
      var after_view = (
        <JSONEdit json={this.props.potential.shaping} onchange={this.setPotential} />
      );
    } else {
      // Shouldn't happen (hopefully)
      var after_view = (
        <b>Something went wrong.</b>
      );
    }
    return (
      <div className="row">
        <div className="col-md-6">
          <div className="row">
            <div className="col-md-6">
              <h4>Current:</h4>
            </div>
            <div className="col-md-6">
              <button type="button" className="btn btn-danger" disabled={!clear_enabled} onClick={this.props.onClearShaping}>
                Clear Shaping
              </button>
            </div>
          </div>
          <div>
            {before_view}
          </div>
        </div>
        <div className="col-md-6">
          <div className="row">
            <div className="col-md-6">
              <h4>New:</h4>
            </div>
            <div className="col-md-3">
              <button type="button" className="btn btn-primary" disabled={this.props.shapingDisabled} onClick={this.props.onPerformShaping}>
                Apply Shaping
              </button>
            </div>
          </div>
          <div>
            {after_view}
          </div>
        </div>
      </div>
    );
  },
});

module.exports = ShapingPanel
