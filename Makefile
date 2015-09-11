# You must have a working go environment in order to build atc.
# See https://golang.org/doc/code.html

PROJECT = github.com/facebook/augmented-traffic-control
SRC = ${PROJECT}/src


TEST = go test -v
BUILD = go build
VET = go vet
FMT = go fmt
THRIFT = thrift

.PHONY: all
all: init bin/atcd bin/atc_api

bin/atcd: src/atc_thrift src/daemon/*.go
	$(FMT) ${SRC}/daemon
	$(VET) ${SRC}/daemon
	$(TEST) ${SRC}/daemon
	$(BUILD) -o bin/atcd ${SRC}/daemon

bin/atc_api: src/atc_thrift src/api/*.go
	$(FMT) ${SRC}/api
	$(VET) ${SRC}/api
	$(TEST) ${SRC}/api
	$(BUILD) -o bin/atc_api ${SRC}/api

src/atc_thrift: if/atc_thrift.thrift
	$(THRIFT) --out src/ --gen go if/atc_thrift.thrift

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
