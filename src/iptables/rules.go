// +build linux

package iptables

import (
	"fmt"
	"strconv"
	"strings"
)

type Rule struct {
	Index       int
	Pkts        int64
	Bytes       int64
	Target      string
	Proto       string
	Opts        string
	In          string
	Out         string
	Source      Target
	Destination Target
	Args        []string
}

func (r Rule) cmdFormat() ([]string, error) {
	if r.Target == "" {
		return nil, fmt.Errorf("Rule has no target.")
	}
	s := make([]string, 0, 5)
	// Only jump is implemented. -g is not used.
	s = append(s, "-j", r.Target)
	if r.Source != nil {
		s = append(s, "-s", r.Source.String())
	}
	if r.Destination != nil {
		s = append(s, "-d", r.Destination.String())
	}
	if r.In != "" {
		s = append(s, "-i", r.In)
	}
	if r.Out != "" {
		s = append(s, "-i", r.Out)
	}
	if r.Proto != "" {
		s = append(s, "-p", r.Proto)
	}
	s = append(s, r.Args...)
	return s, nil
}

func (r Rule) SetMark(mark int64) Rule {
	n := r
	n.Target = "MARK"
	n.Args = append(r.Args, "--set-xmark", fmt.Sprintf("0x%x", mark))
	return n
}

func parseRule(target Target, line string) (*Rule, error) {
	line_tokens := make([]string, 0, 10)
	for _, r := range strings.Split(line, " ") {
		if r != "" {
			line_tokens = append(line_tokens, r)
		}
	}

	m := &Rule{
		Index:       0,
		Pkts:        0,
		Bytes:       0,
		Target:      line_tokens[3],
		Proto:       line_tokens[4],
		Opts:        "",
		In:          "",
		Out:         "",
		Source:      nil,
		Destination: nil,
		Args:        []string{},
	}

	// IPv6 rules can sometimes skip the options field.
	// If the rule has an options field, position 7 will be the source
	// otherwise position 7 will be the output interface (usually '*' for atc)
	// we can determine if the rule has an options field by parsing field 7
	// as an IP
	// all positions are 0-indexed, fyi

	_, err := ParseTarget(line_tokens[7])
	// optPos is the position of the options field, or the position where it would
	// be relative to fields after it.
	// This way we can parse fields by a constant offset after the option position.
	var optPos int
	if err == nil {
		optPos = 4
	} else {
		optPos = 5
		m.Opts = line_tokens[5]
	}

	m.In = line_tokens[optPos+1]
	m.Out = line_tokens[optPos+2]

	m.Source, err = ParseTarget(line_tokens[optPos+3])
	if err != nil {
		return nil, err
	}
	m.Destination, err = ParseTarget(line_tokens[optPos+4])
	if err != nil {
		return nil, err
	}

	if target != nil && m.Source.String() != target.String() && m.Destination.String() != target.String() {
		return nil, nil
	}

	m.Index, err = strconv.Atoi(line_tokens[0])
	if err != nil {
		return nil, fmt.Errorf("Could not parse line number: %v", err)
	}
	m.Pkts, err = strconv.ParseInt(line_tokens[1], 0, 64)
	if err != nil {
		return nil, fmt.Errorf("Could not parse packet count: %v", err)
	}
	m.Bytes, err = strconv.ParseInt(line_tokens[2], 0, 64)
	if err != nil {
		return nil, fmt.Errorf("Could not parse byte count: %v", err)
	}
	if len(line_tokens) > optPos+5 {
		m.Args = line_tokens[optPos+5:]
	}
	return m, nil
}
