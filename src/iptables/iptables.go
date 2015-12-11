// +build linux

package iptables

import (
	"bytes"
	"fmt"
	"os/exec"
)

const (
	default_table = "filter"
)

type IPTables struct {
	Command string
}

func New(command string) *IPTables {
	return &IPTables{command}
}

func (t *IPTables) Table(name string) *Table {
	return &Table{name, t}
}

func (t *IPTables) Chain(name string) *Chain {
	return t.Table(default_table).Chain(name)
}

func (t *IPTables) cmd(argv ...string) (*bytes.Buffer, error) {
	cmd := exec.Command(t.Command, argv...)
	var (
		out_err = &bytes.Buffer{}
		out     = &bytes.Buffer{}
	)
	cmd.Stdout = out
	cmd.Stderr = out_err
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("Could not run %s: %v\n%v", t.Command, err, out_err)
	}
	return out, nil
}
