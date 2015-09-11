# Golang for ATC

## Building

Note: in order to build ATC you must have a working go installation and you
must have `GOPATH` set. See the [Go docs](https://golang.org/doc/code.html)
for details on how to setup a working go environment.

The `Makefile` does most of the heavy lifting.
You should be able to clone `github.com/facebook/augmented-traffic-control`
somewhere and simply run `make`. Generated binaries go into `bin/`.

## Project setup

ATC is split into 2 (eventually probably 3) components:

- `atcd`: the atc daemon
- `atc_api`: the atc REST API
- `atc_ui`: the atc user interface (doesn't exist right now)

Internally the go code is divided into several packages, which encapsulate
different internal sections of the ATC codebase.

The most important of these packages are:

- `atcd`: the executable package for the `atcd` binary
- `atc_api` the executable package for the `atc_api` binary
- `daemon`: the importable package for the common functionality of the atc daemon.
- `api`: the importable package for the common functionality of the atc API.
- `atc_thrift`: generated thrift code following the `atc_thrift.thrift` spec

