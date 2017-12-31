package riotapi

import (
	"fmt"
	"testing"
)

func TestGetMatchesByAccountID(t *testing.T) {
	api := MatchAPI{c: newClient(t)}
	matches, err := api.RecentMatchesByAccountID(29325268)
	if err != nil {
		t.Fatalf("unable to get champions: %v", err)
	}

	if len(matches.Matches) != 20 {
		t.Fatal("invalid recent matches length")
	}
}

func TestGetMatchByID(t *testing.T) {
	api := MatchAPI{c: newClient(t)}
	match, err := api.MatchByID(1858353611)
	if err != nil {
		t.Fatalf("unable to get champions: %v", err)
	}

	fmt.Println(match)
	t.Fail()
}
