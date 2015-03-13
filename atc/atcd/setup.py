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
import re
import sys

from setuptools import setup

readme = open("README.md", "r")

install_requires = [
    'pyroute2==0.3.3',
    'pyotp==1.4.1',
    'sparts==0.7.1',
    'atc_thrift'
]

tests_require = install_requires + [
    'pytest'
]

if sys.version < '3.3':
    tests_require.append('mock')

scripts = ['bin/atcd']


def get_version(package):
    """
    Return package version as listed in `__version__` in `init.py`.
    """
    init_py = open(os.path.join(package, '__init__.py')).read()
    return re.search("__version__ = ['\"]([^'\"]+)['\"]", init_py).group(1)

version = get_version('atcd')

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
    name='atcd',
    version=version,
    description='ATC Daemon',
    author='Emmanuel Bretelle',
    author_email='chantra@fb.com',
    url='https://github.com/facebook/augmented-traffic-control',
    packages=['atcd',
              'atcd.backends',
              'atcd.scripts',
              'atcd.tools'],
    classifiers=['Programming Language :: Python', ],
    long_description=readme.read(),
    scripts=scripts,
    install_requires=install_requires,
    tests_require=tests_require,
)
