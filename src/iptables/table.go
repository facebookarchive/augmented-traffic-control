// +build linux

package iptables

import (
	"bytes"
	"strings"
)

type Table struct {
	Name string
	ipt  *IPTables
}

func (t *Table) Chain(name string) *Chain {
	return &Chain{name, t}
}

func (t *Table) GetRules(target Target) ([]Rule, error) {
	out, err := t.cmd("-nvL", "--line-numbers")
	if err != nil {
		return nil, err
	}

	rules := make([]Rule, 0, 2)
	for i, l := range strings.Split(out.String(), "\n") {
		if i < 2 || strings.TrimSpace(l) == "" {
			continue // skip the headers and empty lines
		}

		if rule, err := parseRule(target, l); err != nil {
			return nil, err
		} else if rule != nil {
			rules = append(rules, *rule)
		}
	}
	return rules, nil
}

func (t *Table) cmd(args ...string) (*bytes.Buffer, error) {
	return t.ipt.cmd(prepend([]string{"-t", t.Name}, args)...)
}

func prepend(before, after []string) []string {
	all := make([]string, 0, len(before)+len(after))
	for _, i := range before {
		all = append(all, i)
	}
	for _, i := range after {
		all = append(all, i)
	}
	return all
}
