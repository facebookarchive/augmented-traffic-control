package main

import (
	"errors"
	"net/url"
	"os"
	"time"

	"github.com/facebook/augmented-traffic-control/src/api"
	logging "github.com/facebook/augmented-traffic-control/src/log"
	"gopkg.in/alecthomas/kingpin.v2"
)

func TestAtcdConnection(thrift_url *url.URL) error {
	atcd := api.NewAtcdConn(thrift_url)
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
	logging.DEBUG = args.Verbose

	// Make sure connection to the daemon is working.
	err := TestAtcdConnection(args.ThriftUrl)
	if err != nil {
		api.Log.Println("failed to connect to atcd server:", err)
		if !args.WarnOnly {
			api.Log.Println("Use --warn if you want to ignore this failure.")
			os.Exit(1)
		}
	} else {
		api.Log.Println("Connected to atcd socket on", args.ThriftUrl)
	}

	if err := validateArgs(args); err != nil {
		api.Log.Fatalln(err)
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

func validateArgs(args Arguments) error {
	if args.V4 == "" && args.V6 == "" {
		return errors.New(
			"You must provide either -4 or -6 arguments to run the API.",
		)
	}

	return nil
}

type Arguments struct {
	api.AtcApiOptions
	WarnOnly bool
	Verbose  bool
}

func ParseArgs() Arguments {
	args := Arguments{}
	kingpin.Flag("listen", "Bind address for the HTTP server").Short('b').Default("0.0.0.0:8080").TCPVar(&args.Addr)
	kingpin.Flag("thrift-addr", "thrift server url (env:ATCD_ADDR)").Short('t').Default("json://127.0.0.1:9090").Envar("ATCD_ADDR").URLVar(&args.ThriftUrl)
	kingpin.Flag("dbdrv", "Database driver").Short('D').Default("sqlite3").StringVar(&args.DBDriver)
	kingpin.Flag("dbconn", "Database connection string").Short('Q').Default("atc_api.db").StringVar(&args.DBConn)
	kingpin.Flag("ipv4", "IPv4 address (or hostname) of the ATC API").Short('4').Default("").StringVar(&args.V4)
	kingpin.Flag("ipv6", "IPv6 address (or hostname) of the ATC API").Short('6').Default("").StringVar(&args.V6)
	kingpin.Flag("proxy-addr", "IP address of authorized HTTP reverse proxy").Default("").StringVar(&args.ProxyAddr)
	kingpin.Flag("warn", "Only warn if the thrift server isn't reachable").Short('Q').Default("false").BoolVar(&args.WarnOnly)
	kingpin.Flag("verbose", "verbose output").Short('v').Default("false").BoolVar(&args.Verbose)
	kingpin.Flag("assets", "location of asset files on disk").Default("").StringVar(&args.AssetPath)
	kingpin.Parse()

	return args
}
