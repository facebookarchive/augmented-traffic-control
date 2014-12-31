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
import sys

from distutils.core import setup

readme = open("README.md", "r")

install_requires = [
    'pyroute2==0.3.3',
    'pyotp==1.4.1',
    'sparts==0.7.1',
    'thrift'
]

tests_require = install_requires + [
    'pytest'
]

if sys.version < '3.3':
    tests_require.append('mock')

scripts = ['bin/atcd']

setup(
    name='atcd',
    version='0.0.1',
    description='ATC Daemon',
    author='Emmanuel Bretelle',
    author_email='chantra@fb.com',
    url='https://github.com/facebook/atc',
    packages=['atcd',
              'atcd.scripts'],
    classifiers=['Programming Language :: Python', ],
    long_description=readme.read(),
    scripts=scripts,
    # FIXME: add atc_thrift dependency once package is published to pip
    install_requires=install_requires,
    tests_require=tests_require,
)
