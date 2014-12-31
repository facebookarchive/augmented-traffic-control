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
import os

from distutils.core import setup

with open(os.path.join(os.path.dirname(__file__), 'README.md')) as readme:
    README = readme.read()


setup(
    name='django-atc-api',
    version='0.0.1',
    description='REST API for ATC',
    author='Emmanuel Bretelle',
    author_email='chantra@fb.com',
    url='https://github.com/facebook/air-traffic-control',
    packages=['atc_api'],
    classifiers=['Programming Language :: Python', ],
    long_description=README,
    install_requires=['atc_thrift', 'djangorestframework']
)
