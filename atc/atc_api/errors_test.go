package main

import (
	"fmt"
	"testing"
)

var e = fmt.Errorf

func TestNoSuchItem(t *testing.T) {
	ShouldMatch := []string{
		"Internal error processing get_group_with: NO_SUCH_ITEM",
	}
	for _, s := range ShouldMatch {
		if !IsNoSuchItem(e(s)) {
			t.Fatal("false negative: %q")
		}
	}
}
