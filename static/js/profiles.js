/** @jsx React.DOM */
/**
 * Copyright (c) 2014, Facebook, Inc.
 * All rights reserved.
 *
 *  This source code is licensed under the BSD-style license found in the
 *  LICENSE file in the root directory of this source tree. An additional grant
 *  of patent rights can be found in the PATENTS file in the same directory.
 */

var ProfilePanel = React.createClass({
  getInitialState: function() {
    return {
      profile_name: "",
    };
  },

  updateProfileName: function(ev) {
    this.setState({profile_name: ev.target.value});
  },

  selectProfile: function(ev) {
    // Ignore index 0 because that's the None option.
    if (ev.target.selectedIndex > 0) {
      // subtract one because the None option.
      profile = this.props.profiles[ev.target.selectedIndex-1];
      // FIXME SETTINGS
      this.props.onSelect(profile.settings);
    }
  },

  saveProfile: function() {
    this.props.onSave(this.state.profile_name);
  },

  render: function() {
    var profilesDisabled = "true";
    var profiles = false;
    if (this.props.profiles != null && this.props.profiles.length > 0) {
      profilesDisabled = "false";
      profiles = this.props.profiles.map(function(item) {
        return (
          <option>{item.name}</option>
        );
      }.bind(this));
    }
    return (
      <div>
        <div className="row">
          <div className="col-md-6">
            <label className="control-label">Profiles:</label>
          </div>
          <div className="col-md-6">
            <select onChange={this.selectProfile} id="profileSelect" className="form-control">
              <option>None</option>
              {profiles}
            </select>
          </div>
        </div>
        <label className="control-label">New Profile Name:</label>
        <div className="row">
          <div className="col-md-6">
            <input type="text" className="form-control" placeholder="name" onChange={this.updateProfileName}/>
          </div>
          <div className="col-md-6">
            <button type="button" className="btn btn-info" onClick={this.saveProfile}>
              Save New As Profile
            </button>
          </div>
        </div>
      </div>
    );
  },
});
