package cli

import (
	"os"
	"testing"
)

func handleValidateArgs(t *testing.T, cmd []string, shouldfail bool) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = cmd
	args := ParseArgs()
	if err := validateArgs(args); err != nil && !shouldfail {
		t.Fatal(err)
	}
}

func TestOnlyIPV4Provided(t *testing.T) {
	handleValidateArgs(
		t,
		[]string{"cmd", "-4", "ipv4"},
		false,
	)
}

func TestOnlyIPV6Provided(t *testing.T) {
	handleValidateArgs(
		t,
		[]string{"cmd", "-6", "ipv6"},
		false,
	)
}

func TestIPV4IPV6Provided(t *testing.T) {
	handleValidateArgs(
		t,
		[]string{"cmd", "-4", "ipv4", "-6", "ipv6"},
		false,
	)
}

func TestNoIPProvided(t *testing.T) {
	handleValidateArgs(t, []string{"cmd"}, true)
}
