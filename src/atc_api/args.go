package main

import (
	"flag"
	"fmt"

	"git.apache.org/thrift.git/lib/go/thrift"
)

const (
	// server bind address
	// overridden by `-b`
	BIND_ADDR = "0.0.0.0:8080"

	// thrift server address
	// overridden by `-t`
	THRIFT_ADDR = "127.0.0.1:9090"
)

// FIXME: globals are nasty
var Args Arguments

type Arguments struct {
	BindAddr       string
	ThriftAddr     string
	ThriftProtocol string
}

var (
	ProtoFactories = map[string]thrift.TProtocolFactory{
		"json": thrift.NewTJSONProtocolFactory(),
	}
)

func (args Arguments) GetThriftProtocol() (thrift.TProtocolFactory, error) {
	f, ok := ProtoFactories[args.ThriftProtocol]
	if !ok {
		return nil, fmt.Errorf("Unknown thrift protocol: %v", args.ThriftProtocol)
	}
	return f, nil
}

func ParseArgs() {
	bindAddr := flag.String("b", BIND_ADDR, "Bind address")
	thriftAddr := flag.String("t", THRIFT_ADDR, "Thrift server address")
	proto := flag.String("p", "json", "Thrift protocol")
	flag.Parse()

	Args = Arguments{
		BindAddr:       *bindAddr,
		ThriftAddr:     *thriftAddr,
		ThriftProtocol: *proto,
	}
}
