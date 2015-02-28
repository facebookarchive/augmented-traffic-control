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
