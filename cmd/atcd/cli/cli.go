// Package cli provides the command line entrypint for the atcd binary
package cli

import (
	"bytes"
	"io"
	"net"
	"os"

	"github.com/facebook/augmented-traffic-control/src/daemon"
	atc_log "github.com/facebook/augmented-traffic-control/src/log"
	"github.com/facebook/augmented-traffic-control/src/shaping"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
	yaml "gopkg.in/yaml.v2"

	"atc_thrift"
	"thrift/lib/go/thrift"
)

func init() {
	daemon.Log = atc_log.NewMux(atc_log.Syslog(), atc_log.Stdlog())
}

func Execute(ox ...OptionFunc) {
	args := parseArgs()
	atc_log.DEBUG = args.Verbose

	// apply functional arguments
	opts := defaults()
	for _, o := range ox {
		opts = o(opts)
	}

	db, err := opts.dbFactory(args.DbDriver, args.DbConnstr)
	if err != nil {
		daemon.Log.Fatalf("Couldn't setup database: %v", err)
	}
	defer db.Close()

	config, err := parseConfig(args.ConfigFile)
	if err != nil {
		if err == os.ErrNotExist {
			daemon.Log.Println("No config file loaded.")
			config = &daemon.Config{}
		} else {
			daemon.Log.Fatalf("Could not parse config file: %v", err)
		}
	}

	// Hooks are called with ATCD_ADDR set to a suitable address
	// for the atc daemon. For this to work, we need to get the address
	// and pass it to the engine
	engine_addr := &net.TCPAddr{
		Port: args.ThriftAddr.Port,
		Zone: args.ThriftAddr.Zone,
	}
	// zero addresses mean listen to everything, so we use the loopback
	if args.ThriftAddr.IP.Equal(net.IPv4zero) {
		engine_addr.IP = net.IPv4(127, 0, 0, 1)
	} else if args.ThriftAddr.IP.Equal(net.IPv6zero) {
		engine_addr.IP = net.IPv6loopback
	} else {
		engine_addr.IP = args.ThriftAddr.IP
	}

	// Wrap the shaper in an engine for hook execution support.
	eng, err := daemon.NewShapingEngine(args.Resolution, engine_addr, config, args.FakeShaping)
	if err != nil {
		daemon.Log.Fatalf("Could not initialize shaping engine: %v", err)
	}

	// Reshape the settings from the database
	if err := daemon.ReshapeFromDb(eng, db); err != nil {
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
		Secure:     !args.Insecure,
		OtpTimeout: uint8(args.OtpTimeout),
	}

	// Create and run the thrift service.
	atcd := daemon.NewAtcd(db, eng, options)
	if err := runServer(atcd, args.ThriftAddr); err != nil {
		daemon.Log.Fatalln("Server failed:", err)
	}
}

type Args struct {
	DbDriver    string
	DbConnstr   string
	ThriftAddr  *net.TCPAddr
	Insecure    bool
	FakeShaping bool
	OtpTimeout  int
	ConfigFile  string
	Verbose     bool
	Resolution  int
}

func parseArgs() Args {
	// ShapingFlags sets up platform-specific flags for the shaper.
	shaping.ShapingFlags()

	args := Args{}

	kingpin.Flag("listen", "Bind address for the thrift server").Short('b').Default("127.0.0.1:9090").TCPVar(&args.ThriftAddr)
	kingpin.Flag("dbdrv", "Database driver").Short('D').Default("sqlite3").StringVar(&args.DbDriver)
	kingpin.Flag("dbconn", "Database connection string").Short('Q').Default("atcd.db").StringVar(&args.DbConnstr)
	kingpin.Flag("config", "location of the json config file").Short('c').Default("/etc/atc/atcd.conf").StringVar(&args.ConfigFile)

	// flag is `insecure` because security is the default and you should have
	// to turn it off deliberately
	// note that this means we're using a double negative. be careful what you
	// change here.
	kingpin.Flag("insecure", "insecure mode. disable user security checks").Default("false").BoolVar(&args.Insecure)
	kingpin.Flag("fake-shaping", "don't do real shaping. instead use a mock shaper").Short('F').Default("false").BoolVar(&args.FakeShaping)
	kingpin.Flag("token-timeout", "OTP Token timeout in seconds").Default("60").IntVar(&args.OtpTimeout)
	kingpin.Flag("verbose", "verbose output").Short('v').Default("false").BoolVar(&args.Verbose)
	kingpin.Flag("resolution", "periodic hook resolution in seconds").Default("5").IntVar(&args.Resolution)

	kingpin.Parse()

	return args
}

// Runs the ATCD thrift server on the provided address.
func runServer(atcd atc_thrift.Atcd, addr *net.TCPAddr) error {
	transport, err := thrift.NewTServerSocket(addr.String())
	if err != nil {
		return err
	}
	processor := atc_thrift.NewAtcdProcessor(atcd)

	pfactory := thrift.NewTJSONProtocolFactory()
	tfactory := thrift.NewTTransportFactory()
	server := thrift.NewTSimpleServer4(processor, transport, tfactory, pfactory)

	daemon.Log.Printf("Starting the thrift server on %v\n", addr)
	return server.Serve()
}

func parseConfig(filename string) (*daemon.Config, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, os.ErrNotExist
	}
	defer f.Close()
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, f); err != nil {
		return nil, err
	}
	var config daemon.Config
	if err := yaml.Unmarshal(buf.Bytes(), &config); err != nil {
		return nil, err
	}
	return &config, nil
}
