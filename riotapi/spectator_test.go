package riotapi

import (
	"fmt"
	"testing"
)

func TestActiveGamesBySummonerID(t *testing.T) {
	api := SpectatorAPI{newClient(t)}
	ag, err := api.ActiveGamesBySummoner(24749077)
	if err != nil {
		t.Fatalf("unable to get active game: %v", err)
	}

	fmt.Println(ag)
	t.Fail()
}
