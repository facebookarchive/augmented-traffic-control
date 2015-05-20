/**
 * Copyright (c) 2014, Facebook, Inc.
 * All rights reserved.
 *
 *  This source code is licensed under the BSD-style license found in the
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
                /*
                console.log('API status: ' + status);
                if (status == 'success' || status == 'nocontent' || xhr.status == 404) {
                    if (status == 'success') {
                        rc.settings = new AtcSettings().mergeWithDefaultSettings({
                            up: xhr.responseJSON.up,
                            down: xhr.responseJSON.down,
                        });
                    } else {
                        rc.settings = new AtcSettings().getDefaultSettings();
                    }
                } else {
                    rc.detail = xhr.responseJSON.detail;
                }
                console.log(rc);
                */
                if (callback !== undefined) {
                    callback(rc);
                }
            }
        });
    };
}


AtcRestClient.prototype.shape = function (callback, data) {
    this.api_call('POST', 'shape', callback, data);
};

AtcRestClient.prototype.unshape = function (callback, data) {
    this.api_call('DELETE', 'shape', callback);
};

AtcRestClient.prototype.getCurrentShaping = function (callback) {
    this.api_call('GET', 'shape', callback);
};

AtcRestClient.prototype.getToken = function (callback) {
    this.api_call('GET', 'token', callback);
};

AtcRestClient.prototype.getAuthInfo = function (callback) {
    this.api_call('GET', 'auth', callback);
};

AtcRestClient.prototype.updateAuthInfo = function (address, data, callback) {
    this.api_call('POST', 'auth/'.concat(address), callback, data);
};

function AtcSettings () {
    this.defaults = {
        'up': {
            'rate': null,
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
        'down': {
            'rate': null,
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

    this.getDefaultSettings = function () {
        return $.extend(true, {}, this.defaults);
    };

    this.mergeWithDefaultSettings = function (data) {
        return $.extend(true, {}, this.defaults, data);
    };
}
