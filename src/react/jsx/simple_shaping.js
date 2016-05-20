/**
 * Copyright (c) 2014, Facebook, Inc.
 * All rights reserved.
 *
 *  This source code is licensed under the BSD-style license found in the
 *  LICENSE file in the root directory of this source tree. An additional grant
 *  of patent rights can be found in the PATENTS file in the same directory.
 */

var React = require('react');

var SimpleShapingPanel = React.createClass({
  getInitialState: function() {
    return {
      profile_name: "",
    };
  },

  selectProfile: function(ev) {
    // Ignore index 0 because that's the None option.
    var profile = null;
    if (ev.target.selectedIndex > 0) {
      // subtract one because the None option.
      profile = this.props.profiles[ev.target.selectedIndex-1];
    }
    this.props.onSelectProfile(profile != null ? profile.shaping : null);
  },


  render: function() {
    var profiles = false;
    if (this.props.profiles != null && this.props.profiles.length > 0) {
      profiles = this.props.profiles.map(function(item) {
        return (
          <option key={item.id}>{item.name}</option>
        );
      }.bind(this));
    }

    var shaping_button_style = 'btn-primary';
    var shaping_button_text = 'Apply Shaping';
    if (this.props.shaped) {
      shaping_button_style = 'btn-danger';
      shaping_button_text = 'Clear Shaping';
    }
    return (
      <div className="row">
        <div className="col-md-6">
            <select onChange={this.selectProfile} id="profileSelect" className="form-control" disabled={this.props.shaped}>
                <option>None</option>
                {profiles}
            </select>
        </div>
        <div className="col-md-6">
            <button type="button" className={'btn ' + shaping_button_style} onClick={this.props.onToggleShaping} disabled={!this.props.shaped && this.props.potential_shaping == null} >
              {shaping_button_text}
            </button>
        </div>
      </div>
    );
  },
});

module.exports = SimpleShapingPanel
