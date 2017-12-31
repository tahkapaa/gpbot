package riotapi

import (
	"fmt"
	"net/http"
)

// SpectatorAPI implements the Riot Spectator API methods
type SpectatorAPI struct {
	c *Client
}

const spectatorAPIPath = "spectator"

// CurrentGameInfo represents a game in progress
type CurrentGameInfo struct {
	GameID            int
	GameStartTime     int
	PlatformID        string
	GameMode          string
	MapID             int
	GameType          string
	BannedChampions   []BannedChampion
	Observers         Observer
	Participants      []CurrentGameParticipant
	GameLength        int
	GameQueueConfigID int
}

// BannedChampion contains information about banned champions
type BannedChampion struct {
	PickTurn   int
	ChampionID int
	TeamID     int
}

// Observer information
type Observer struct {
	EncryptionKey string
}

// CurrentGameParticipant contains information about a participant
type CurrentGameParticipant struct {
	ProfileIconID            int
	ChampionID               int
	SummonerName             string
	GameCustomizationObjects []GameCustomizationObject
	Bot                      bool
	Perks                    Perks
	Spell1ID                 int
	Spell2ID                 int
	TeamID                   int
	SummonerID               int
}

// GameCustomizationObject contains information about game customization
type GameCustomizationObject struct {
	Category string
	Content  string
}

// Perks / runes reforged information
type Perks struct {
	PerkStyle    int
	PerkIDs      []int
	PerkSubStyle int
}

// ActiveGamesBySummoner gets current game information for the given summoner ID
// or nil, if there is no game
func (api SpectatorAPI) ActiveGamesBySummoner(id int) (*CurrentGameInfo, error) {
	var cgi CurrentGameInfo
	if err := api.c.Request(spectatorAPIPath, fmt.Sprintf("active-games/by-summoner/%d", id), &cgi); err != nil {
		if apiErr, ok := err.(APIError); ok {
			if apiErr.StatusCode == http.StatusNotFound {
				return nil, nil
			}
		}
		return nil, err
	}
	return &cgi, nil
}
