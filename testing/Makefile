#
#  Copyright (c) 2014, Facebook, Inc.
#  All rights reserved.
#
#  This source code is licensed under the BSD-style license found in the
#  LICENSE file in the root directory of this source tree. An additional grant
#  of patent rights can be found in the PATENTS file in the same directory.
#
#
VMS = gateway client server


.PHONY: default
default: vup test vdown


.PHONY: test
test:
	nosetests -s e2e/


.PHONY: vup
vup:
	VAGRANT_CWD=./e2e/ vagrant up ${VMS}


.PHONY: vdown
vdown:
	VAGRANT_CWD=./e2e/ vagrant destroy -f ${VMS}


.PHONY: lint
lint:
	flake8 e2e/
