// +build linux

package iptables

import (
	"fmt"
	"strings"
)

type Chain struct {
	Name  string
	table *Table
}

func (c *Chain) Append(r Rule) error {
	f, err := r.cmdFormat()
	if err != nil {
		return fmt.Errorf("Could not append to %s chain: %v", c.Name, err)
	}
	if _, err := c.table.cmd(prepend([]string{"-A", c.Name}, f)...); err != nil {
		return fmt.Errorf("Could not append to %s chain: %v", c.Name, err)
	} else {
		return nil
	}
}

func (c *Chain) Flush() error {
	if _, err := c.table.cmd("-F", c.Name); err != nil {
		return fmt.Errorf("Could not flush %s chain: %v", c.Name, err)
	} else {
		return nil
	}
}

func (c *Chain) Delete(r Rule) error {
	f, err := r.cmdFormat()
	if err != nil {
		return fmt.Errorf("Could not delete from %s chain: %v", c.Name, err)
	}
	if _, err := c.table.cmd(prepend([]string{"-D", c.Name}, f)...); err != nil {
		return fmt.Errorf("Could not delete from %s chain: %v", c.Name, err)
	} else {
		return nil
	}
}

func (c *Chain) GetRules(target Target) ([]Rule, error) {
	out, err := c.table.cmd("-nvL", c.Name, "--line-numbers")
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
