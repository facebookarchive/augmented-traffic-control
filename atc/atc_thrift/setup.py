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
import sys
from setuptools import setup

version = '0.1.3'

readme = open('README.md', 'r')


if sys.argv[-1] == 'publish':
    if os.system("pip freeze | grep wheel"):
        print("wheel not installed.\nUse `pip install wheel`.\nExiting.")
        sys.exit()
    if os.system("pip freeze | grep twine"):
        print("twine not installed.\nUse `pip install twine`.\nExiting.")
        sys.exit()
    os.system("python setup.py sdist bdist_wheel")
    os.system("twine upload dist/*")
    print("You probably want to also tag the version now:")
    print("  git tag -a %s -m 'version %s'" % (version, version))
    print("  git push --tags")
    sys.exit()

setup(
    name='atc_thrift',
    version=version,
    description='ATC Thrift Library',
    author='Emmanuel Bretelle',
    author_email='chantra@fb.com',
    url='https://github.com/facebook/augmented-traffic-control',
    packages=['atc_thrift'],
    classifiers=['Programming Language :: Python', ],
    long_description=readme.read(),
    install_requires=['thrift']
)
