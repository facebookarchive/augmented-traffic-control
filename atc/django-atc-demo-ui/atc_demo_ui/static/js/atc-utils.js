/** @jsx React.DOM */
/**
 * Copyright (c) 2014, Facebook, Inc.
 * All rights reserved.
 *
 *  This source code is licensed under the BSD-style license found in the
 *  LICENSE file in the root directory of this source tree. An additional grant
 *  of patent rights can be found in the PATENTS file in the same directory.
 */


var JSONView = React.createClass({
  render: function() {
    return (
      <div className="col-md-6">
      <h4>{this.props.label}</h4>
      <pre>
        { JSON.stringify(this.props.json, null, 2) }
      </pre>
      </div>
    );
  }
});


function capitalizeFirstLetter(s) {
  return s.charAt(0).toUpperCase() + s.slice(1);
}


/** https://gist.github.com/NV/8622188 **/
/**
 * RecursiveLinkStateMixin is a LinkState alternative that can update keys in
 * a dictionnary recursively.
 * You can either give it a string of keys separated by a underscore (_)
 * or a list of keys
 */
var RecursiveLinkStateMixin = {
  linkState: function (path) {
    function setPath (obj, path, value) {
      var leaf = resolvePath(obj, path);
      leaf.obj[leaf.key] = value;
    }

    function getPath (obj, path) {
      var leaf = resolvePath(obj, path);
      return leaf.obj[leaf.key];
    }

    function resolvePath (obj, keys) {
      if (typeof keys === 'string') {
        keys = keys.split('_');
      }
      var lastIndex = keys.length - 1;
      var current = obj;
      for (var i = 0; i < lastIndex; i++) {
        var key = keys[i];
        current = current[key];
      }
      return {
        obj: current,
        key: keys[lastIndex]
      };
    }

    return {
      value: getPath(this.state, path),
      requestChange: function(newValue) {
        setPath(this.state, path, newValue);
        this.forceUpdate();
      }.bind(this)
    };
  }
};


var IdentifyableObject = {
  getIdentifier: function () {
    return this.props.params.join('_');
  },
};


function handleAPI(callback, notify) {
  return function(rc) {
    /* 2XX error codes are OK */
    if (rc.status < 300 && rc.status >= 200) {
      if (callback !== undefined) {
        callback(rc);
      }  
    } else {
      err = false;
      t = typeof rc.json;
      if (t === 'undefined') {
        err = "Could not complete request due to server error."
      } else if (t === 'string') {
        s = rc.json;

        /* trim off the first line */
        s = s.trim().substring(s.length, s.indexOf('\n'));

        /* take the second line as the error message */
        s = s.trim().substring(0, s.indexOf('\n'));

        err = s;
      } else if (t === 'object') {
        err = rc.json.detail;
      }

      if (err) {
        notify('error', err);
      } else {
        console.log("Not sure what to do with error value " + t + " '" + rc.json + "'");
      }
    }
  }
}