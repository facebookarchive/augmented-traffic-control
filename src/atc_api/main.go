package main

import (
	"net"
	"os"
	"time"

	"github.com/facebook/augmented-traffic-control/src/api"
	"gopkg.in/alecthomas/kingpin.v2"
)

func TestAtcdConnection(addr *net.TCPAddr, proto string) error {
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
	args := Arguments{}
	kingpin.Flag("listen", "Bind address for the HTTP server").Short('b').Default("0.0.0.0:8080").TCPVar(&args.Addr)
	kingpin.Flag("atcd", "ATCD thrift server address").Short('t').Default("127.0.0.1:9090").TCPVar(&args.ThriftAddr)
	kingpin.Flag("atcd-proto", "ATCD thrift server protocol").Short('p').Default("json").StringVar(&args.ThriftProto)
	kingpin.Flag("dbdrv", "Database driver").Short('D').Default("sqlite3").StringVar(&args.DBDriver)
	kingpin.Flag("dbconn", "Database connection string").Short('Q').Default("atc_api.db").StringVar(&args.DBConn)
	kingpin.Flag("ipv4", "IPv4 address (or hostname) of the ATC API").Short('4').Default("").StringVar(&args.DBConn)
	kingpin.Flag("ipv6", "IPv6 address (or hostname) of the ATC API").Short('6').Default("").StringVar(&args.DBConn)
	kingpin.Flag("proxy-addr", "IP address of authorized HTTP reverse proxy").Default("").StringVar(&args.ProxyAddr)
	kingpin.Flag("warn", "Only warn if the thrift server isn't reachable").Short('Q').Default("false").BoolVar(&args.WarnOnly)
	kingpin.Parse()

	return args
}
