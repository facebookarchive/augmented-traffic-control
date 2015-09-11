# You must have a working go environment in order to build atc.
# See https://golang.org/doc/code.html

PROJECT = github.com/facebook/augmented-traffic-control
SRC = ${PROJECT}/src

BINARIES = bin/atcd bin/atc_api

TEST = go test -v
BUILD = go build
THRIFT = thrift

.PHONY: all
all: init $(BINARIES)

bin/atcd: src/atc_thrift src/atcd/*.go
	$(TEST) ${SRC}/atcd
	$(BUILD) -o bin/atcd ${SRC}/atcd

bin/atc_api: src/atc_thrift src/atc_api/*.go
	$(TEST) ${SRC}/atc_api
	$(BUILD) -o bin/atc_api ${SRC}/atc_api

src/atc_thrift: src/atc_thrift.thrift
	$(THRIFT) --out src/ --gen go src/atc_thrift.thrift

.PHONY: init
init:
	mkdir -p bin/
	mkdir -p ${GOPATH}/src/${PROJECT}
	[ -h ${GOPATH}/src/${SRC} ] || ln -s $(shell pwd)/src ${GOPATH}/src/${SRC}

.PHONY: clean
clean:
	rm -rf bin/

.PHONY: fullclean
fullclean: clean
	rm -rf src/atc_thrift/
