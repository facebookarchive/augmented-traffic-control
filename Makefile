# You must have a working go environment in order to build atc.
# See https://golang.org/doc/code.html


# Path to install files under
PREFIX = /usr/local

# for static compilation:
#BUILD = go build --ldflags '-extldflags "-static"'
GO = $(shell which go)
BUILD = $(GO) build

TEST = $(GO) test -v
VET = $(GO) vet
FMT = $(GO) fmt
GET = $(GO) get
LIST = $(GO) list
BINGEN = $(GOPATH)/bin/go-bindata # github.com/jteeuwen/go-bindata
THRIFT = thrift

# The $(GO) project root
PROJECT = github.com/facebook/augmented-traffic-control
SRC = ${PROJECT}/src

STATIC_FILES = $(shell find static/ -print)

USERID = $(shell id -u)

.PHONY: all bin
all: bin
bin: bin/atcd bin/atc_api bin/atc

bin/atcd: src/daemon/*.go src/atcd/*.go src/log/*.go src/shaping/*.go
	@$(FMT) ${SRC}/shaping ${SRC}/daemon ${SRC}/atcd
	@$(VET) ${SRC}/shaping ${SRC}/daemon ${SRC}/atcd
	@mkdir -p bin
	$(BUILD) -o $@ ${SRC}/atcd

bin/atc_api: src/api/bindata.go src/api/*.go src/atc_api/*.go src/log/*.go
	@$(FMT) ${SRC}/api ${SRC}/atc_api
	@$(VET) ${SRC}/api ${SRC}/atc_api
	@mkdir -p bin
	$(BUILD) -o $@ ${SRC}/atc_api

bin/atc: src/log/*.go src/atc/*.go
	@$(FMT) ${SRC}/atc
	@$(VET) ${SRC}/atc
	@mkdir -p bin
	$(BUILD) -o $@ ${SRC}/atc

.PHONY: tests
tests: src/api/bindata.go
	$(TEST) ${SRC}/daemon
	@echo "[31mRunning shaping tests as root.[39m"
ifeq ($(USERID),0)
	$(TEST) ${SRC}/shaping
else
	sudo PATH=${PATH} GOROOT=${GOROOT} GOPATH=${GOPATH} $(TEST) ${SRC}/shaping
endif
	$(TEST) ${SRC}/atcd
	$(TEST) ${SRC}/api
	$(TEST) ${SRC}/atc_api

src/api/bindata.go: $(STATIC_FILES)
	$(BINGEN) -pkg api -o $@ static/...

src/atc_thrift: if/atc_thrift.thrift
	$(THRIFT) --out src/ --gen $(GO) if/atc_thrift.thrift

# Removed compiled binaries
.PHONY: clean
clean:
	rm -rf bin/
	rm -f src/api/bindata.go

# Copy built binaries into /usr/local/bin/
.PHONY: install
install:
	cp bin/atcd bin/atc_api "$(PREFIX)/bin/"
