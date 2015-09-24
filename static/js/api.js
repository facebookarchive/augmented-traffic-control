/** @jsx React.DOM */
/**
 * Copyright (c) 2014, Facebook, Inc.
 * All rights reserved.  * *  This source code is licensed under the BSD-style license found in the
 *  LICENSE file in the root directory of this source tree. An additional grant
 *  of patent rights can be found in the PATENTS file in the same directory.
 */


function AtcRestClient (endpoint) {
  this.endpoint = endpoint || '/api/v1/';
  function _add_ending_slash(string) {
    if (string[string.length -1] != '/') {
      string += '/';
    }
    return string;
  }

  this.endpoint = _add_ending_slash(this.endpoint);

  this.api_call = function (method, urn, callback, data) {
    urn = _add_ending_slash(urn);
    $.ajax({
      url: this.endpoint + urn,
      dataType: 'json',
      type: method,
      data: data && JSON.stringify(data),
      contentType: 'application/json; charset=utf-8',
      complete: function (xhr, status) {
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

AtcRestClient.prototype.joinGroup = function(id, token, callback) {
  this.api_call('POST', 'group/' + id.toString() + '/join', callback, token)
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

function defaultSettings() {
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
