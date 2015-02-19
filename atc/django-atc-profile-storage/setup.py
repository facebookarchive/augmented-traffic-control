#!/usr/bin/env python
import os

from distutils.core import setup

with open(os.path.join(os.path.dirname(__file__), 'README.md')) as readme:
    README = readme.read()


setup(
    name='django-atc-profile-storage',
    version='0.0.1',
    description='ATC Profile storage app',
    author='Emmanuel Bretelle',
    author_email='chantra@fb.com',
    url='https://github.com/facebook/air-traffic-control',
    packages=['atc_profile_storage'],
    classifiers=['Programming Language :: Python', ],
    long_description=README,
    install_requires=['djangorestframework']
)
