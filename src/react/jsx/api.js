/**
 * Copyright (c) 2014, Facebook, Inc.
 * All rights reserved.  * *  This source code is licensed under the BSD-style license found in the
 *  LICENSE file in the root directory of this source tree. An additional grant
 *  of patent rights can be found in the PATENTS file in the same directory.
 */


function AtcRestClient (endpoint) {
  this.endpoint = endpoint || '/api/v1/';
  this.addresses = {'primary': "", 'secondary': ""};

  function _add_ending_slash(string) {
    if (string[string.length -1] != '/') {
      string += '/';
    }
    return string;
  }

  this.endpoint = _add_ending_slash(this.endpoint);

  this.dual_stack = function() {
    return this.addresses['secondary'] != "";
  }

  this.raw_call = function (addr, method, urn, callback, data) {
    /**
    * If addr is empty, default to using the one from the url we connected to.
    * Also wrap ipv6 with square brackets and set a sane default port.
    */
    if (addr == "") {
      addr = document.location["hostname"];
    }
    var port = document.location["port"];
    if (addr.indexOf(':') >= 0) {
      addr = '[' + addr + ']';
    }
    urn = _add_ending_slash(urn);
    $.ajax({
      url: '//' + addr + (port != "" ? ":" + port : "") + this.endpoint + urn,
      dataType: 'json',
      type: method,
      data: data && JSON.stringify(data),
      contentType: 'application/json; charset=utf-8',
      complete: function (xhr, status) { //eslint-disable-line no-unused-vars
        var rc = {
          status: xhr.status,
          json: xhr.responseJSON,
        };
        if (callback !== undefined) {
          callback(rc);
        }
      }
    });
  };

  this.api_call = function (method, urn, callback, data) {
    this.raw_call(this.addresses['primary'], method, urn, callback, data);
  };

  this.secondary_call = function (method, urn, callback, data) {
    this.raw_call(this.addresses['secondary'], method, urn, callback, data);
  };

  function discover_addresses(rc) {
    var c = rc['json']['client'];
    this.addresses['primary'] = c['server_primary'];
    this.addresses['secondary'] = c['server_secondary'];
  }

  this.api_call('GET', 'info', discover_addresses.bind(this));
}

AtcRestClient.prototype.getServerInfo = function (callback) {
  this.api_call('GET', 'info', callback)
}

AtcRestClient.prototype.getGroup = function (callback) {
  this.api_call('GET', 'group', callback)
}

AtcRestClient.prototype.createGroup = function(callback) {
  this.api_call('POST', 'group', callback)
}

AtcRestClient.prototype.leaveGroup = function(id, token, callback) {
  this.api_call('POST', 'group/' + id.toString() + '/leave', callback, token)
}

AtcRestClient.prototype.leaveGroupSecondary = function(id, token, callback) {
  this.secondary_call('POST', 'group/' + id.toString() + '/leave', callback, token)
}

AtcRestClient.prototype.joinGroup = function(id, token, callback) {
  this.api_call('POST', 'group/' + id.toString() + '/join', callback, token)
}

AtcRestClient.prototype.joinGroupSecondary = function(id, token, callback) {
  this.secondary_call('POST', 'group/' + id.toString() + '/join', callback, token)
}

AtcRestClient.prototype.getToken = function(id, callback) {
  this.api_call('GET', 'group/' + id.toString() + '/token', callback)
}

AtcRestClient.prototype.getShaping = function (callback) {
  this.api_call('GET', 'shape', callback);
};

AtcRestClient.prototype.shape = function (data, callback) {
  this.api_call('POST', 'shape', callback, data);
};

AtcRestClient.prototype.unshape = function (callback) {
  this.api_call('DELETE', 'shape', callback);
};

AtcRestClient.prototype.getProfiles = function (callback) {
  this.api_call('GET', 'profile', callback);
};

AtcRestClient.prototype.createProfile = function(profile, callback) {
  this.api_call('POST', 'profile', callback, profile)
}

AtcRestClient.prototype.defaultShaping = function () {
  return {
    'down': {
      'rate': 100,
      'delay': {
        'delay': 0,
        'jitter': 0,
        'correlation': 0
      },
      'loss': {
        'percentage': 0,
        'correlation': 0
      },
      'reorder': {
        'percentage': 0,
        'correlation': 0,
        'gap': 0
      },
      'corruption': {
        'percentage': 0,
        'correlation': 0
      },
      'iptables_options': Array(),
    },
    'up': {
      'rate': 10,
      'delay': {
        'delay': 0,
        'jitter': 0,
        'correlation': 0
      },
      'loss': {
        'percentage': 0,
        'correlation': 0
      },
      'reorder': {
        'percentage': 0,
        'correlation': 0,
        'gap': 0
      },
      'corruption': {
        'percentage': 0,
        'correlation': 0
      },
      'iptables_options': Array(),
    }
  };
}

module.exports = AtcRestClient
