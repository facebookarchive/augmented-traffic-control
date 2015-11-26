/**
 * Copyright (c) 2014, Facebook, Inc.
 * All rights reserved.  * *  This source code is licensed under the BSD-style license found in the
 *  LICENSE file in the root directory of this source tree. An additional grant
 *  of patent rights can be found in the PATENTS file in the same directory.
 */

var React = require('react');
var ReactDOM = require('react-dom');
var Atc = require('./atc.js');

ReactDOM.render(
    <Atc primary={primary} secondary={secondary} endpoint={endpoint} />,
    document.getElementById('atc_demo_ui')
);
