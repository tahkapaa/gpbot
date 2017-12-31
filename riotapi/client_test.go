package riotapi

import "testing"

const aPIKey = ""

func newClient(t *testing.T) *Client {
	c, err := New(aPIKey, "eune", 50, 20)
	if err != nil {
		t.Fatalf("unable to create client: %v", err)
	}
	return c
}
