package main

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/tahkapaa/gangplankbot/riotapi"
)

var participantsJSON = `{
    "gameId": 1868762103,
    "gameStartTime": 0,
    "platformId": "EUN1",
    "gameMode": "CLASSIC",
    "mapId": 11,
    "gameType": "CUSTOM_GAME",
    "bannedChampions": [],
    "observers": {
        "encryptionKey": "L9tTaGJ7DHmOZa10rJT9tr6afpiSBm3f"
    },
    "participants": [
        {
            "profileIconId": 3232,
            "championId": 201,
            "summonerName": "Uxipaxa",
            "gameCustomizationObjects": [],
            "bot": false,
            "perks": {
                "perkStyle": 8300,
                "perkIds": [
                    8359,
                    8345,
                    8304,
                    8347,
                    8451,
                    8430
                ],
                "perkSubStyle": 8400
            },
            "spell2Id": 12,
            "teamId": 100,
            "spell1Id": 4,
            "summonerId": 24749077
        }
    ],
    "gameLength": 0
}`

const DgToken = ""

func TestPlayerMonitor_showRunes(t *testing.T) {
	DDC = riotapi.NewDDragonClient()
	c, err := riotapi.New(apiKey, "eune", 50, 20)
	if err != nil {
		t.Errorf("unable to initialize riot api: %v", err)
	}
	RC = c

	bot, err := newBot(DgToken, New("test"))
	if err != nil {
		t.Errorf("Unable to create bot: %v\n", err)
	}
	DGBot = bot

	g := make(map[int]*riotapi.CurrentGameInfo)
	var cgi riotapi.CurrentGameInfo
	if err := json.Unmarshal([]byte(participantsJSON), &cgi); err != nil {
		t.Errorf("unable to unmarshall json: %v", err)
	}
	g[0] = &cgi

	pm := PlayerMonitor{
		lastGameID: 0,
		games:      g,
		ChannelID:  "386887549836328964",
	}

	pm.showRunes(0, "Testi", &discordgo.User{Username: "Aikaleima"})
	DGBot.Discord.Close()
}

func TestEmojis(t *testing.T) {
	bot, err := newBot(DgToken, New("test"))
	if err != nil {
		t.Errorf("Unable to create bot: %v\n", err)
	}
	guild, err := bot.Discord.Guild("386887549408247826")
	if err != nil {
		t.Errorf("failed to get guild: %v", err)
	}

	for _, emo := range guild.Emojis {
		fmt.Println("ID: ", emo.ID, "Name: ", emo.Name, emo.APIName())
	}
	t.Fail()
}
