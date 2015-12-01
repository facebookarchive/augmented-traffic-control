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

	// Make sure connection to the daemon is working.
	err := TestAtcdConnection(args.ThriftAddr, args.ThriftProto)
	if err != nil {
		api.Log.Println("failed to connect to atcd server:", err)
		if !args.WarnOnly {
			os.Exit(1)
		}
	} else {
		api.Log.Println("Connected to atcd socket on", args.ThriftAddr)
	}

	if args.V4 == "" && args.V6 == "" {
		api.Log.Fatalln("You must provide either -4 or -6 arguments to run the API.")
	}

	api.Log.Println("Listening on", args.Addr)

	srv, err := api.ListenAndServe(args.AtcApiOptions)
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
	api.AtcApiOptions
	WarnOnly bool
}

func ParseArgs() Arguments {
	bindAddr := flag.String("b", "0.0.0.0:8080", "Bind address")
	thriftAddr := flag.String("t", "127.0.0.1:9090", "Thrift server address")
	proto := flag.String("p", "json", "Thrift protocol")
	db_driver := flag.String("D", "sqlite3", "database driver")
	db_connstr := flag.String("Q", "atc_api.db", "database driver connection parameters")
	warn_only := flag.Bool("W", false, "only warn if the thrift server isn't reachable")
	ipv4 := flag.String("4", "", "IPv4 address for the API")
	ipv6 := flag.String("6", "", "IPv6 address for the API")
	flag.Parse()

	return Arguments{
		AtcApiOptions: api.AtcApiOptions{
			Addr:        *bindAddr,
			ThriftAddr:  *thriftAddr,
			ThriftProto: *proto,
			DBDriver:    *db_driver,
			DBConn:      *db_connstr,
			V4:          *ipv4,
			V6:          *ipv6,
		},
		WarnOnly: *warn_only,
	}
}
