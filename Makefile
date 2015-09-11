# You must have a working go environment in order to build atc.
# See https://golang.org/doc/code.html

PROJECT = github.com/facebook/augmented-traffic-control
SRC = ${PROJECT}/src

BINARIES = bin/atcd bin/atc_api

TEST = go test -v
BUILD = go build
THRIFT = thrift

.PHONY: all
all: init tests $(BINARIES)

.PHONY: tests
tests:
	$(TEST) ${SRC}/atcd
	$(TEST) ${SRC}/atc_api

bin/atcd: src/atc_thrift src/atcd/*.go
	$(BUILD) -o bin/atcd ${SRC}/atcd

bin/atc_api: src/atc_thrift src/atc_api/*.go
	$(BUILD) -o bin/atcd ${SRC}/atc_api

src/atc_thrift: init src/atc_thrift.thrift
	$(THRIFT) --out src/ --gen go src/atc_thrift.thrift

.PHONY: init
init:
	mkdir -p bin/
	# This symlink is required for the go build commands to work.
	mkdir -p ${GOPATH}/src/${PROJECT}
	[ -d ${GOPATH}/src/${PROJECT}/src ] || ln -s $(shell pwd)/src ${GOPATH}/src/${PROJECT}/src

.PHONY: clean
clean:
	rm -rf bin/

.PHONY: fullclean
fullclean: clean
	rm -rf src/atc_thrift/
