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


# By default, we just lint since we want this to be **fast**.
# In the future when unit testing becomes better this should run quick tests.
.PHONY: default
default: lint


# Do all the things!
.PHONY: all
all: lint fulltest


# Lint the various sources that ATC includes:
#  chef/  - chef cookbooks
#  atc/   - ATC source code
#  tests/ - ATC test code
.PHONY: lint
lint:
	rubocop chef/atc
	foodcritic chef/atc
	pep8 atc
	flake8 atc
	flake8 tests/


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
