package api

import (
	"fmt"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/facebook/augmented-traffic-control/src/atc_thrift"
)

var (
	ProtoFactories = map[string]thrift.TProtocolFactory{
		"json": thrift.NewTJSONProtocolFactory(),
	}
)

func getThriftProtocol(thrift_proto string) (thrift.TProtocolFactory, error) {
	f, ok := ProtoFactories[thrift_proto]
	if !ok {
		return nil, fmt.Errorf("Unknown thrift protocol: %q", thrift_proto)
	}
	return f, nil
}

type AtcdConn struct {
	*atc_thrift.AtcdClient
	xport        thrift.TTransport
	thrift_addr  string
	thrift_proto string
}

func NewAtcdConn(thrift_addr, thrift_proto string) *AtcdConn {
	return &AtcdConn{
		thrift_addr:  thrift_addr,
		thrift_proto: thrift_proto,
	}
}

func (atcd *AtcdConn) Open() error {
	var err error
	atcd.xport, err = thrift.NewTSocket(atcd.thrift_addr)
	if err != nil {
		return err
	}
	if err := atcd.xport.Open(); err != nil {
		return err
	}

	pfactory, err := getThriftProtocol(atcd.thrift_proto)
	if err != nil {
		return err
	}
	atcd.AtcdClient = atc_thrift.NewAtcdClientFactory(atcd.xport, pfactory)
	return nil
}

func (atcd *AtcdConn) Close() {
	atcd.xport.Close()
}
