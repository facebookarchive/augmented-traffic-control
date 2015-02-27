VMS = gateway client server


.PHONY: default
default: vup test vdown


.PHONY: test
test:
	nosetests -s e2e/


.PHONY: vup
vup:
	vagrant up ${VMS}


.PHONY: vdown
vdown:
	vagrant destroy -f ${VMS}


.PHONY: lint
lint:
	flake8 e2e/
