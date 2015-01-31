# nosetests will search for files named test*.py
# since vagrant.py doesn't follow this pattern,
# it needs help to be found by the test runner
from e2e.vagrant import tearDownModule, setUpModule
