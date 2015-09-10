package main

import (
	"testing"

	"github.com/facebook/augmented-traffic-control/atc/atc_api/stub"
	"github.com/facebook/augmented-traffic-control/atc/atc_thrift"
)

var (
	atcd atc_thrift.Atcd = stub.NewFakeAtcd()
)

func Cleanup(t *Server) {
	t.Atcd.Close()
	t.Atcd = nil
	t.Kill()
	return
}

func StubServer() *Server {
	srv := &Server{
		Atcd: nil,
	}
	return srv
}

func TestGetGroup(t *testing.T) {
	t.Skip("Need to rewrite")
	srv := StubServer()
	defer Cleanup(srv)
	data_, err := GroupHandler(atcd, nil, FakeRequest("GET", "/group/15", nil))
	if err != nil {
		t.Fatal(err)
	}
	group := data_.(*atc_thrift.ShapingGroup)
	if group.Id != stub.FakeGroup.Id {
		t.Errorf("Wrong group id expected %d != %d", stub.FakeGroup.Id, group.Id)
	}
}
