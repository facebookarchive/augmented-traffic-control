package main

import (
	"flag"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/facebook/augmented-traffic-control/src/atc_thrift"
	"github.com/facebook/augmented-traffic-control/src/daemon"
)

func main() {
	args := parseArgs()
	db, err := daemon.NewDbRunner(args.DbDriver, args.DbConnstr)
	if err != nil {
		daemon.Log.Fatalf("Couldn't setup database: %v", err)
	}
	var shaper daemon.Shaper
	if !args.FakeShaping {
		shaper, err = daemon.GetShaper()
		if err != nil {
			daemon.Log.Fatalf("Couldn't get shaper: %v", err)
		}
	} else {
		daemon.Log.Println("Using fake shaper. Your network isn't actually being shaped!")
		shaper = daemon.FakeShaper{}
	}
	defer db.Close()

	if args.OtpTimeout > 255 {
		daemon.Log.Println("Can't use token timeout >255. Setting to 255s")
		args.OtpTimeout = 255
	}
	options := &daemon.AtcdOptions{
		Secure:     args.Secure,
		OtpTimeout: uint8(args.OtpTimeout),
	}
	atcd := daemon.NewAtcd(db, shaper, options)
	runServer(atcd, args.ThriftAddr)
}

type Args struct {
	DbDriver    string
	DbConnstr   string
	ThriftAddr  string
	Secure      bool
	FakeShaping bool
	OtpTimeout  int
}

func parseArgs() Args {
	db_driver := flag.String("D", "sqlite3", "database driver")
	db_connstr := flag.String("Q", "atcd.db", "database driver connection parameters")
	thrift_addr := flag.String("B", "127.0.0.1:9090", "bind address for the thrift server")
	// flag is `insecure` because security is the default and you should have
	// to turn it off deliberately
	insecure := flag.Bool("I", false, "disable secure mode")
	fake_shaping := flag.Bool("F", false, "don't do real shaping. instead use a mock shaper")
	otp_timeout := flag.Int("token-timeout", 60, "Token timeout in seconds")

	flag.Parse()

	return Args{
		DbDriver:    *db_driver,
		DbConnstr:   *db_connstr,
		ThriftAddr:  *thrift_addr,
		Secure:      !*insecure,
		FakeShaping: *fake_shaping,
		OtpTimeout:  *otp_timeout,
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

	daemon.Log.Println("Starting the thrift server on", addr)
	return server.Serve()
}
