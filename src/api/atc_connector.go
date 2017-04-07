package api

import (
	"fmt"
	"net/url"

	"github.com/apache/thrift/lib/go/thrift"
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

type AtcdCloser interface {
	atc_thrift.Atcd
	Close()
}

type AtcdConn struct {
	*atc_thrift.AtcdClient
	xport      thrift.TTransport
	thrift_url *url.URL
}

func NewAtcdConn(thrift_url *url.URL) *AtcdConn {
	return &AtcdConn{
		thrift_url: thrift_url,
	}
}

func (atcd *AtcdConn) Open() error {
	var err error
	atcd.xport, err = thrift.NewTSocket(atcd.thrift_url.Host)
	if err != nil {
		return err
	}
	if err := atcd.xport.Open(); err != nil {
		return err
	}

	pfactory, err := getThriftProtocol(atcd.thrift_url.Scheme)
	if err != nil {
		return err
	}
	atcd.AtcdClient = atc_thrift.NewAtcdClientFactory(atcd.xport, pfactory)
	return nil
}

func (atcd *AtcdConn) Close() {
	atcd.xport.Close()
}
