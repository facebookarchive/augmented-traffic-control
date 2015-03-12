/**
 * Copyright (c) 2014, Facebook, Inc.
 * All rights reserved.
 *
 *  This source code is licensed under the BSD-style license found in the
 *  LICENSE file in the root directory of this source tree. An additional grant
 *  of patent rights can be found in the PATENTS file in the same directory.
 */

AtcRestClient.prototype.new_profile = function (callback, data) {
    this.api_call('POST', 'profiles', callback, data);
};

AtcRestClient.prototype.delete_profile = function (callback, id) {
    this.api_call('DELETE', 'profiles/' + id.toString() + '/', callback);
};

AtcRestClient.prototype.get_profile = function (callback, id) {
    this.api_call('GET', 'profiles/' + id.toString() + '/', callback);
};

AtcRestClient.prototype.get_profiles = function (callback) {
    this.api_call('GET', 'profiles/', callback);
};
