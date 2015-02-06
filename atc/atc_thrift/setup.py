#!/usr/bin/env python
#
#  Copyright (c) 2014, Facebook, Inc.
#  All rights reserved.
#
#  This source code is licensed under the BSD-style license found in the
#  LICENSE file in the root directory of this source tree. An additional grant
#  of patent rights can be found in the PATENTS file in the same directory.
#
#

from distutils.core import setup

readme = open("README.md", "r")

setup(
    name='atc_thrift',
    version='0.0.1',
    description='ATC Thrift Library',
    author='Emmanuel Bretelle',
    author_email='chantra@fb.com',
    url='https://github.com/facebook/augmented-traffic-control',
    packages=['atc_thrift'],
    classifiers=['Programming Language :: Python', ],
    long_description=readme.read(),
    install_requires=['thrift']
)
