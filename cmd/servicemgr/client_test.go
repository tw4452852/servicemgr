package main

import (
	"testing"
)

func TestClientId(t *testing.T) {
	c1 := NewClient(nil)
	c2 := NewClient(nil)
	if got1, got2 := c1.Id(), c2.Id(); got2 != got1+1 {
		t.Fatalf("second client's id[%d] isn't the first one[%d]+1\n", got2, got1)
	}
}
