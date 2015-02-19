lint:
	rubocop chef/atc
	foodcritic chef/atc
	pep8 atc
	flake8 atc
	flake8 testing/

test:
	make -C testing/ test
