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
    var notifyNodes = this.props.notifications.map(function(item, idx, arr) {
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
    if (notifyNodes.length == 0) {
      notifyNodes = (
        <i>No notifications.</i>
      );
    }
    return (
      <div className="panel-group" id="accordionNotify" role="tablist" aria-multiselectable="false">
        <div className="panel panel-default">
          <div className="panel-heading" data-toggle="collapse" data-parent="#accordionNotify" href="#collapseNotify" aria-expanded="true" aria-controls="collapseNotify">
            <h3 className="panel-title">
              Notifications
            </h3>
          </div>
          <div id="collapseNotify" className="panel-collapse collapse in" role="tabpanel">
            <div className="panel-body">
              {notifyNodes}
            </div>
          </div>
        </div>
      </div>
    );
  }
});


var Atc = React.createClass({
  mixins: [RecursiveLinkStateMixin],
  getInitialState: function() {
    return {
      client: new AtcRestClient(this.props.endpoint),
      settings: new AtcSettings().getDefaultSettings(),
      current_settings: new AtcSettings().getDefaultSettings(),
      status: atc_status.OFFLINE,
      profiles: [],
      notifications: [],
    };
  },

  notify: function(type, msg) {
    this.setState(function(state, props) {
      return {
        notifications: state.notifications.concat({
          expire_at: ERROR_EXPIRY + new Date().getTime(),
          message: msg,
          type: type,
        })
      };
    });
  },

  expireNotifications: function() {
    this.setState(function(state, props) {
      return {
        notifications: state.notifications.filter(function(item, idx, arr) {
            return item.expire_at >= new Date().getTime();
        })
      };
    })
  },

  componentDidMount: function() {
    this.getCurrentShaping();
    /** FIXME we are calling getCurrentShaping to make sure that
     * current_settings === settings.... let's be smarter than that.
     */
    this.getCurrentShaping();
    this.getProfiles();
    this.expiry_interval = setInterval(this.expireNotifications, 1000);
  },

  componentWillUnmount: function() {
    if (this.expiry_interval != null) {
      clearInterval(this.expiry_interval);
    }
  },

  handleClick: function(e) {
    if (e.type == "click") {
      if (this.state.status == atc_status.ACTIVE) {
        this.unsetShaping();
      } else if (this.state.status == atc_status.INACTIVE) {
        this.setShaping();
      }
    }
  },

  updateClick: function(e) {
    if (e.type == "click") {
      this.setShaping();
    }
  },

  hasChanged: function() {
    /* TODO: improve object comparaison e.g null == "", 1 == "1"*/
    function objectEquals(x, y) {
      if (typeof(x) === 'number') {
        x = x.toString();
      }
      if (typeof(y) === 'number') {
        y = y.toString();
      }
      if (typeof(x) != typeof(y)) {
        return false;
      }

      if (Array.isArray(x) || Array.isArray(y)) {
        return x.toString() === y.toString();
      }

      if (x === null && y === null) {
        return true;
      }

      if (typeof(x) === 'object' && x !== null) {
        x_keys = Object.keys(x);
        y_keys = Object.keys(y);
        if (x_keys.sort().toString() !== y_keys.sort().toString()) {
          console.error('Object do not have the same keys: ' +
            x_keys.sort().toString() + ' vs ' +
            y_keys.sort().toString()
          );
          return false;
        }
        equals = true;
        x_keys.forEach(function (key, index) {
            equals &= objectEquals(x[key], y[key]);
        });
        return equals;
      }
      return x.toString() === y.toString();
    }
    return !objectEquals(this.state.settings, this.state.current_settings);
  },

  getProfiles: function() {
    this.state.client.get_profiles(function (result) {
      if (result.status >= 200 && result.status < 300) {
        this.setState({
          profiles: result.json,
        });
      } else {
        this.error(result.json.detail);
        this.setState({
          profiles: [],
        });
      }
    }.bind(this));
  },

  getCurrentShaping: function() {
    this.state.client.getCurrentShaping(function (result) {
      if (result.status == 404) {
        this.setState({
          status: atc_status.INACTIVE,
          settings: new AtcSettings().getDefaultSettings(),
          current_settings: new AtcSettings().getDefaultSettings(),
        });
      } else if (result.status >= 200 && result.status < 300) {
        this.setState({
          status: atc_status.ACTIVE,
          settings: result.json,
          current_settings: this.state.settings,
        });
      } else {
        this.error(result.json.detail);
        this.setState({
          status: atc_status.OFFLINE,
          settings: new AtcSettings().getDefaultSettings(),
        });
      }
    }.bind(this));
  },

  unsetShaping: function() {
    console.log('unsetShaping');
    this.state.client.unshape(function (result) {
      if (result.status >= 200 && result.status < 300) {
        this.setState({
          status: atc_status.INACTIVE,
          settings: new AtcSettings().getDefaultSettings(),
          current_settings: new AtcSettings().getDefaultSettings(),
        });
      } else if (result.status >= 500) {
        this.notify("error", result.json.detail);
        this.setState({
          status: atc_status.OFFLINE,
        });
      }
    }.bind(this));
  },


  setShaping: function() {
    console.log('setShaping');
    this.state.client.shape(function (result) {
      if (result.status >= 200 && result.status < 300) {
        this.setState({
          status: atc_status.ACTIVE,
          settings: result.json,
          current_settings: {down: this.state.settings.down, up: this.state.settings.up},
        });
      } else if (result.status == 400) {
        for (var key in result.json) {
          result.data[key].map(function(msg) {
            this.notify("error", key + ': ' + msg);
          }.bind(this));
        }
      } else if (result.status >= 500) {
        this.notify("error", result.json.detail);
        this.setState({
          status: atc_status.OFFLINE,
        });
      }

    }.bind(this), {down: this.state.settings.down, up: this.state.settings.up});
  },

  render: function () {
    link_state = this.linkState;
    var err_msg = "";
    var update_button = false;
    if (this.hasChanged()) {
      update_button = <ShapingButton id="update_button" status={atc_status.OUTDATED} onClick={this.updateClick} />
    }
    return (
      <div>
        <div className="row">
          <div id="shaping_buttons" className="col-md-12 text-center">
            {update_button}
            <ShapingButton id="shaping_button" status={this.state.status} onClick={this.handleClick} />
            {err_msg}
          </div>
        </div>

        <NotificationPanel notifications={this.state.notifications} />
        <AuthPanel client={this.state.client} notify={this.notify}/>
        <ProfilePanel refreshProfiles={this.getProfiles} link_state={link_state} profiles={this.state.profiles} notify={this.notify} />
        <ShapingSettings link_state={link_state} before={this.state.settings} after={this.state.current_settings} notify={this.notify} />
      </div>
    )
  }
});
