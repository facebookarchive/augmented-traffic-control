/** @jsx React.DOM */
/**
 * Copyright (c) 2014, Facebook, Inc.
 * All rights reserved.
 *
 *  This source code is licensed under the BSD-style license found in the
 *  LICENSE file in the root directory of this source tree. An additional grant
 *  of patent rights can be found in the PATENTS file in the same directory.
 */


var Profile = React.createClass({
  getInitialState: function() {
    return {
      name: "",
    };
  },

  handleClick: function() {
    this.props.link_state("settings").requestChange(
      new AtcSettings().mergeWithDefaultSettings(this.props.profile.content)
    );
  },

  updateName: function(event) {
    this.setState({name: event.target.value});
  },

  removeProfile: function() {
    this.props.link_state("client").value.delete_profile(handleAPI(this.props.refreshProfiles, this.props.notify), this.props.profile.id);
  },

  render: function () {
    return (
      <div className="list-group-item row">
        <span className="col-sm-6 text-center vcenter"><kbd>{this.props.profile.name}</kbd></span>
        <span className="col-sm-2 text-center vcenter">{this.props.profile.content.up.rate} kbps</span>
        <span className="col-sm-2 text-center vcenter">{this.props.profile.content.down.rate} kbps</span>
        <button className="col-sm-1 btn btn-info vcenter" onClick={this.handleClick}>Select</button>
        <button className="col-sm-1 btn btn-danger vcenter" onClick={this.removeProfile}>Delete</button>
      </div>);
  }
});


var ProfileList = React.createClass({
  render: function() {
    if (this.props.profiles.length == 0) {
      return false;
    }

    var profileNodes = this.props.profiles.map(function (profile) {
      return (
        <Profile refreshProfiles={this.props.refreshProfiles} link_state={this.props.link_state} action='delete' profile={profile} notify={this.props.notify} />
      );
    }.bind(this));

    return (
      <div>
        <h4>Existing Profiles</h4>
        <p>
          Select a profile from the list below to use it.
        </p>
        <div className="list-group">
          <div className="list-group-item row">
            <span className="col-sm-6 text-center vcenter"><b>Name</b></span>
            <span className="col-sm-2 text-center vcenter"><b>Up Rate</b></span>
            <span className="col-sm-2 text-center vcenter"><b>Down Rate</b></span>
            <span className="col-sm-1 text-center vcenter"></span>
            <span className="col-sm-1 text-center vcenter"></span>
          </div>

          {profileNodes}
        </div>
      </div>
    );
  }
});


var CreateProfileWidget = React.createClass({
  getInitialState: function() {
    return {
      name: ""
    };
  },

  updateName: function(event) {
    this.setState({name: event.target.value});
  },

  newProfile: function() {
    var failed = false;
    var settings = this.props.link_state('settings').value;
    if (settings.down.rate == null &&
      settings.up.rate == null) {
      this.props.notify("error", "You must enter shaping settings below.");
      failed = true;
    }
    if (this.state.name == "") {
      this.props.notify("error", "You must give the new profile a name.");
      failed = true;
    }
    if (failed) {
      return;
    }

    var addProfile = function() {
      this.setState({
        name: "",
      });
      this.props.refreshProfiles();
    }.bind(this);

    var profile = {
      name: this.state.name,
      content: settings
    };
    this.props.link_state("client").value.new_profile(handleAPI(addProfile, this.props.notify), profile);
  },

  render: function() {
    return (
      <div>
        <h4>New Profile</h4>
        <p>
          Enter a name and click "Create" to save a new profile with the settings under "Shaping Settings" below.
        </p>
        <input type="text" className="form-control" placeholder="Profile Name" onChange={this.updateName}/>
        <button className="col-sm-2 btn btn-success" onClick={this.newProfile}>Create</button>
      </div>
    );
  },
});


var ProfilePanel = React.createClass({
  render: function () {
    return (
      <div className="panel-group" id="accordion1" role="tablist" aria-multiselectable="false">
        <div className="panel panel-default">
          <div className="panel-heading" data-toggle="collapse" data-parent="#accordion1" href="#collapseProfiles" aria-expanded="false" aria-controls="collapseProfiles">
            <h3 className="panel-title">
              Profiles
            </h3>
          </div>
          <div id="collapseProfiles" className="panel-collapse collapse" role="tabpanel">
            <div className="panel-body">
              <ProfileList refreshProfiles={this.props.refreshProfiles} link_state={this.props.link_state} profiles={this.props.profiles} notify={this.props.notify}/>

              <CreateProfileWidget refreshProfiles={this.props.refreshProfiles} link_state={this.props.link_state} notify={this.props.notify}/>
            </div>
          </div>
        </div>
      </div>
    );
  }
});