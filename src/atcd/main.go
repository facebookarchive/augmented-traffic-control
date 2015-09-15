package main

import (
	"flag"
	"fmt"
	"log"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/facebook/augmented-traffic-control/src/atc_thrift"
	"github.com/facebook/augmented-traffic-control/src/daemon"
)

func main() {
	args := parseArgs()
	db, err := daemon.NewDbRunner(args.DbDriver, args.DbConnstr)
	if err != nil {
		log.Fatalf("Couldn't setup database: %v", err)
	}
	defer db.Close()
	atcd := daemon.NewAtcd(db, daemon.GetShaper(), args.Secure)
	runServer(atcd, "127.0.0.1:9090")
}

type Args struct {
	DbDriver   string
	DbConnstr  string
	ThriftAddr string
	Secure     bool
}

func parseArgs() Args {
	// Letters here mostly chosen arbitrarily
	db_driver := flag.String("D", "sqlite3", "database driver")
	// fixme change to actual file
	db_connstr := flag.String("Q", "atcd.db", "database driver connection parameters")
	thrift_addr := flag.String("B", "127.0.0.1:9090", "bind address for the thrift server")
	insecure := flag.Bool("I", false, "disable secure mode")

	flag.Parse()

	return Args{
		DbDriver:   *db_driver,
		DbConnstr:  *db_connstr,
		ThriftAddr: *thrift_addr,
		Secure:     !*insecure,
	}
}

// Runs the ATCD thrift server on the provided address.
func runServer(atcd atc_thrift.Atcd, addr string) error {
	transport, err := thrift.NewTServerSocket(addr)
	if err != nil {
		return err
	}
	processor := atc_thrift.NewAtcdProcessor(atcd)

	pfactory := thrift.NewTJSONProtocolFactory()
	tfactory := thrift.NewTTransportFactory()
	server := thrift.NewTSimpleServer4(processor, transport, tfactory, pfactory)

	fmt.Println("Starting the thrift server on", addr)
	return server.Serve()
}
