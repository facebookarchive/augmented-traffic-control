VMS = gateway client server

.PHONY: default
default: lint

.PHONY: all
all: lint fulltest

.PHONY: lint
lint:
	rubocop chef/atc
	foodcritic chef/atc
	pep8 atc
	flake8 atc
	flake8 tests/


.PHONY: fulltest
fulltest: vup test vdown


.PHONY: test
test:
	nosetests -s tests/


.PHONY: vup
vup:
	VAGRANT_CWD=./tests/ vagrant up ${VMS}


.PHONY: vdown
vdown:
	VAGRANT_CWD=./tests/ vagrant destroy -f ${VMS}
