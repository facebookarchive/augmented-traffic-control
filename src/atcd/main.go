package main

import (
	"flag"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/facebook/augmented-traffic-control/src/atc_thrift"
	"github.com/facebook/augmented-traffic-control/src/daemon"
	"github.com/facebook/augmented-traffic-control/src/shaping"
)

func main() {
	args := parseArgs()

	// Setup the database
	db, err := daemon.NewDbRunner(args.DbDriver, args.DbConnstr)
	if err != nil {
		daemon.Log.Fatalf("Couldn't setup database: %v", err)
	}
	defer db.Close()

	// Setup the shaper
	var shaper shaping.Shaper
	if !args.FakeShaping {
		shaper, err = shaping.GetShaper()
		if err != nil {
			daemon.Log.Fatalf("Couldn't get shaper: %v", err)
		}
	} else {
		daemon.Log.Println("Using fake shaper. Your network isn't actually being shaped!")
		shaper = shaping.FakeShaper{}
	}
	if err := shaper.Initialize(); err != nil {
		daemon.Log.Fatalf("Could not initialize shaping: %v", err)
	}

	// Reshape the settings from the database
	if err := daemon.ReshapeFromDb(shaper, db); err != nil {
		// The reason we do this is to ensure that settings from previous atcd
		// instances are preserved across restarts/reboots or network outages.
		// We can't gaurantee that the settings in the shaper are the same as those
		// in the database, so we reset the shaper settings on startup.
		daemon.Log.Fatalf("Could not reshape settings from database: %v", err)
	}

	// Setup options
	if args.OtpTimeout > 255 {
		// The OTP library uses an int8 for their token timeouts.
		// If the user specifies a value > 255, this will cause an overflow
		// error later, so we prevent them from setting it that high.
		daemon.Log.Println("Can't use token timeout >255. Setting to 255 seconds")
		args.OtpTimeout = 255
	}
	options := &daemon.AtcdOptions{
		Secure:     args.Secure,
		OtpTimeout: uint8(args.OtpTimeout),
	}

	// Create and run the thrift service.
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
	// ShapingFlags sets up platform-specific flags for the shaper.
	shaping.ShapingFlags()
	db_driver := flag.String("D", "sqlite3", "database driver")
	db_connstr := flag.String("Q", "atcd.db", "database driver connection parameters")
	thrift_addr := flag.String("B", "127.0.0.1:9090", "bind address for the thrift server")
	// flag is `insecure` because security is the default and you should have
	// to turn it off deliberately
	// note that this means we're using a double negative. be careful what you
	// change here.
	insecure := flag.Bool("insecure", false, "insecure mode. disable user security checks")
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
