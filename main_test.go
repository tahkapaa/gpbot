package main

import (
	"testing"

	"github.com/tahkapaa/gangplankbot/riotapi"
)

var apiKey = ""

func TestFindPlayerRank(t *testing.T) {
	c, err := riotapi.New(apiKey, "eune", 50, 20)
	if err != nil {
		t.Fatal(err)
	}
	s, err := findPlayerRank(c, 29652836)
	if err != nil {
		t.Fatal(err)
	}

	if s == "Not found" {
		t.Errorf("invalid response: %s", s)
	}
}
