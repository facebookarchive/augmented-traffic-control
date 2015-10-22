# You must have a working go environment in order to build atc.
# See https://golang.org/doc/code.html


# Path to install files under
PREFIX = /usr/local

# for static compilation:
#BUILD = go build --ldflags '-extldflags "-static"'
BUILD = go build

TEST = go test -v
VET = go vet
FMT = go fmt
GET = go get
LIST = go list
BINGEN = $(GOPATH)/bin/go-bindata # github.com/jteeuwen/go-bindata
THRIFT = thrift

# The go project root
PROJECT = github.com/facebook/augmented-traffic-control
SRC = ${PROJECT}/src

STATIC_FILES = $(shell find static/ -print)

.PHONY: all
all: bin/atcd bin/atc_api

bin/atcd: src/daemon/*.go src/atcd/*.go src/log/* src/shaping/*.go
	@$(FMT) ${SRC}/shaping ${SRC}/daemon ${SRC}/atcd
	@$(VET) ${SRC}/shaping ${SRC}/daemon ${SRC}/atcd
	$(TEST) ${SRC}/daemon
	@echo "[31mRunning shaping tests as root.[39m"
	sudo GOPATH=${GOPATH} $(TEST) ${SRC}/shaping
	$(TEST) ${SRC}/atcd
	@mkdir -p bin
	$(BUILD) -o $@ ${SRC}/atcd

bin/atc_api: src/api/bindata.go src/api/*.go src/atc_api/*.go src/log/*
	@$(FMT) ${SRC}/api ${SRC}/atc_api
	@$(VET) ${SRC}/api ${SRC}/atc_api
	$(TEST) ${SRC}/api
	$(TEST) ${SRC}/atc_api
	@mkdir -p bin
	$(BUILD) -o $@ ${SRC}/atc_api

src/api/bindata.go: $(STATIC_FILES)
	$(BINGEN) -pkg api -o $@ static/...

src/atc_thrift: if/atc_thrift.thrift
	$(THRIFT) --out src/ --gen go if/atc_thrift.thrift

# Removed compiled binaries
.PHONY: clean
clean:
	rm -rf bin/
	rm -f src/api/bindata.go

# Copy built binaries into /usr/local/bin/
.PHONY: install
install:
	cp bin/atcd bin/atc_api "$(PREFIX)/bin/"
