

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
