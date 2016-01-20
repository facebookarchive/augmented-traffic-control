#
#  Copyright (c) 2014, Facebook, Inc.
#  All rights reserved.
#
#  This source code is licensed under the BSD-style license found in the
#  LICENSE file in the root directory of this source tree. An additional grant
#  of patent rights can be found in the PATENTS file in the same directory.
#


# Vagrant names used for testing.
# Described fully in `tests/Vagrantfile`.
TESTVMS = gateway client server

FLAKE=python -m flake8
PEP=pep8


# By default, we just lint since we want this to be **fast**.
# In the future when unit testing becomes better this should run quick tests.
.PHONY: default
default: lint


# Do all the things!
.PHONY: all
all: lint fulltest


# Install packages locally.
.PHONY: install
install:
	cd atc/atc_thrift && pip install --upgrade --force-reinstall .
	cd atc/atcd && pip install --upgrade --force-reinstall .
	cd atc/django-atc-api && pip install --upgrade --force-reinstall .
	cd atc/django-atc-demo-ui && pip install --upgrade --force-reinstall .
	cd atc/django-atc-profile-storage && pip install --upgrade --force-reinstall .


# Publish packages to PyPi.
.PHONY: publish
publish:
	cd atc/atc_thrift && python setup.py publish
	cd atc/atcd && python setup.py publish
	cd atc/django-atc-api && python setup.py publish
	cd atc/django-atc-demo-ui && python setup.py publish
	cd atc/django-atc-profile-storage && python setup.py publish


# Cleans up python dist files
.PHONY: clean
clean:
	rm -rf atc/{atcd,atc_thrift,django-atc-api,django-atc-demo-ui,django-atc-profile-storage}/{build,dist}


# Lint the various sources that ATC includes:
#  chef/  - chef cookbooks
#  atc/   - ATC source code
#  tests/ - ATC test code
.PHONY: lint
lint: chef_lint python_lint

.PHONY: chef_lint
chef_lint:
	rubocop chef/atc
	foodcritic chef/atc

.PHONY: python_lint
python_lint:
	$(PEP) atc
	$(FLAKE) atc
	$(FLAKE) tests/


# Performs setup, runs the tests, then cleans up.
# Should be used for automated testing.
.PHONY: fulltest
fulltest: testvup test testvdown


# Runs the ATC test suite.
# This can be run manually for quick testing.
# Requires that the test VMs have been created by `testvup`
.PHONY: test
test:
	nosetests -s tests/


# Creates vagrant VMs for testing.
.PHONY: testvup
testvup:
	cd tests/ && vagrant up ${TESTVMS}


# Tears down vagrant VMs.
.PHONY: testvdown
testvdown:
	cd tests/ && vagrant destroy -f ${TESTVMS}
