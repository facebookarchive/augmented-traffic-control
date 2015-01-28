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

install_requires = [
    'django-atc-api',
    'django-static-jquery==1.11.1',
    'django-bootstrap-themes==3.1.2',
]

setup(
    name='django-atc-demo-ui',
    version='0.0.1',
    description='Demo Web UI for ATC',
    author='Emmanuel Bretelle',
    author_email='chantra@fb.com',
    url='https://github.com/facebook/air-traffic-control',
    packages=['atc_demo_ui'],
    classifiers=['Programming Language :: Python', ],
    long_description=README,
    install_requires=install_requires,
)
