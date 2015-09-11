package main

import (
	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/facebook/augmented-traffic-control/src/atc_thrift"
)

type AtcdConn struct {
	*atc_thrift.AtcdClient
	xport thrift.TTransport
}

func NewAtcdConn() *AtcdConn {
	return &AtcdConn{}
}

func (atcd *AtcdConn) Open() error {
	var err error
	atcd.xport, err = thrift.NewTSocket(Args.ThriftAddr)
	if err != nil {
		return err
	}
	if err := atcd.xport.Open(); err != nil {
		return err
	}

	pfactory, err := Args.GetThriftProtocol()
	if err != nil {
		return err
	}
	atcd.AtcdClient = atc_thrift.NewAtcdClientFactory(atcd.xport, pfactory)
	return nil
}

func (atcd *AtcdConn) Close() {
	atcd.xport.Close()
}
