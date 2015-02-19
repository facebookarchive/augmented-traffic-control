VMS = gateway client server

# Uncomment to disable cleaning up vagrant VMs after tests.
# For repeated testing, it is recommended that you do this, since
# it significantly cuts down on the test time (from ~7m to ~2m).
NOCLEAN=yes

.PHONY: test
test: e2e

.PHONY: e2e
e2e: vup te2e vdown

.PHONY: te2e
te2e:
	# runtime: ~100s
	nosetests -s e2e/

.PHONY: vup
vup:
	# runtime: ~5m
	vagrant up ${VMS}

.PHONY: vdown
vdown:
	# runtime: ~20s
ifndef NOCLEAN
	vagrant destroy -f ${VMS}
else
	@echo "Not cleaning up vagrant VMs."
endif

lint:
	flake8 e2e/
