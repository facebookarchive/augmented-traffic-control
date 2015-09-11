package main

import (
	"fmt"
	"os"
	"time"
)

const (
	// Current version
	VERSION = "2.0-go"
)

func GetApiInfo() APIInfo {
	return APIInfo{
		Version: VERSION,
	}
}

func TestAtcdConnection() error {
	atcd := NewAtcdConn()
	if err := atcd.Open(); err != nil {
		return err
	}
	if _, err := atcd.GetAtcdInfo(); err != nil {
		return err
	}
	return nil
}

func main() {
	ParseArgs()

	if err := TestAtcdConnection(); err != nil {
		fmt.Fprintln(os.Stderr, "failed to connect to atcd server:", err)
		os.Exit(1)
	}

	fmt.Fprintln(os.Stderr, "Connected to atcd socket on", Args.ThriftAddr)
	fmt.Fprintln(os.Stderr, "Listening on", Args.BindAddr)

	srv, err := ListenAndServe(Args.BindAddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to listen and serve:", err)
		os.Exit(1)
	}
	defer srv.Kill()
	for {
		// Let the server run
		time.Sleep(100 * time.Second)
	}
}
