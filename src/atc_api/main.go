package main

import (
	"flag"
	"os"
	"time"

	"github.com/facebook/augmented-traffic-control/src/api"
)

func TestAtcdConnection(addr, proto string) error {
	atcd := api.NewAtcdConn(addr, proto)
	if err := atcd.Open(); err != nil {
		return err
	}
	if _, err := atcd.GetAtcdInfo(); err != nil {
		return err
	}
	return nil
}

func main() {
	args := ParseArgs()

	err := TestAtcdConnection(args.ThriftAddr, args.ThriftProtocol)
	if err != nil {
		api.Log.Println("failed to connect to atcd server:", err)
		if !args.WarnOnly {
			os.Exit(1)
		}
	} else {
		api.Log.Println("Connected to atcd socket on", args.ThriftAddr)
	}

	api.Log.Println("Listening on", args.BindAddr)

	srv, err := api.ListenAndServe(args.BindAddr, args.ThriftAddr, args.ThriftProtocol, args.DbDriver, args.DbConnstr)
	if err != nil {
		api.Log.Fatalln("failed to listen and serve:", err)
	}
	defer srv.Kill()
	for {
		// Let the server run
		time.Sleep(100 * time.Second)
	}
}

type Arguments struct {
	BindAddr       string
	ThriftAddr     string
	ThriftProtocol string
	DbDriver       string
	DbConnstr      string
	WarnOnly       bool
}

func ParseArgs() Arguments {
	bindAddr := flag.String("b", "0.0.0.0:8080", "Bind address")
	thriftAddr := flag.String("t", "127.0.0.1:9090", "Thrift server address")
	proto := flag.String("p", "json", "Thrift protocol")
	db_driver := flag.String("D", "sqlite3", "database driver")
	db_connstr := flag.String("Q", ":memory:", "database driver connection parameters")
	warn_only := flag.Bool("W", false, "only warn if the thrift server isn't reachable")
	flag.Parse()

	return Arguments{
		BindAddr:       *bindAddr,
		ThriftAddr:     *thriftAddr,
		ThriftProtocol: *proto,
		DbDriver:       *db_driver,
		DbConnstr:      *db_connstr,
		WarnOnly:       *warn_only,
	}
}
