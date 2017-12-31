package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/tahkapaa/gangplankbot/riotapi"
)

// PlayerMonitor monitors players
type PlayerMonitor struct {
	FollowedPlayers map[string]Player
	games           map[int]*riotapi.CurrentGameInfo
	reportedGames   map[int]bool
	messageChan     chan monitorMessage
	ChannelID       string
	region          string
	db              DB
	lastGameID      int
	ReportRunes     bool
	RuneDesc        string
}

const (
	// RuneShort shows a short written description
	RuneShort = "short"
	// RuneLong shows a longer written description
	RuneLong = "long"
)

type monitorMessage struct {
	mtype        msgType
	player       Player
	summonerName string
	playerIndex  int
	runeDesc     string
	author       discordgo.User
	timeStr      string
	region       string
}

type msgType int

const (
	AddPlayer msgType = iota
	RemovePlayer
	ListPlayers
	ShowGame
	ShowRunes
	ToggleRunesReporting
	SetRegion
	SetRuneDesc
)

const (
	teamNone = 0
	teamRed  = 200
	teamBlue = 100
)

func (pm *PlayerMonitor) save() {
	if err := pm.db.Save(&ChannelData{ID: pm.ChannelID, Region: pm.region, Summoners: pm.FollowedPlayers, ReportRunes: pm.ReportRunes, RuneDesc: pm.RuneDesc}); err != nil {
		log.Println("Unable to save channel data")
	}
}

func (pm *PlayerMonitor) monitorPlayers() {
	for {
		select {
		case msg := <-pm.messageChan:
			switch msg.mtype {
			case AddPlayer:
				log.Printf("added player '%v' to channel '%v'\n", msg.player.Name, pm.ChannelID)
				pm.FollowedPlayers[strconv.Itoa(msg.player.ID)] = msg.player
				pm.save()
			case RemovePlayer:
				log.Printf("'%v' from channel: %v\n", msg.summonerName, pm.ChannelID)
				pm.removePlayer(&msg)
			case ListPlayers:
				log.Printf("list players for channel: %v\n", pm.ChannelID)
				pm.listPlayers(msg.timeStr, &msg.author)
			case ShowGame:
				log.Printf("show game for '%v' in channel: %v\n", msg.summonerName, pm.ChannelID)
				pm.showGame(&msg)
			case ShowRunes:
				log.Printf("show runes for player: %v in channel: %v\n", msg.playerIndex, pm.ChannelID)
				pm.showRunes(msg.playerIndex, msg.timeStr, &msg.author)
			case ToggleRunesReporting:
				log.Printf("toggle runes for channel: %v\n", pm.ChannelID)
				pm.toggleRunes(msg.timeStr, &msg.author)
			case SetRegion:
				log.Printf("set region '%v' for channel: %v\n", msg.region, pm.ChannelID)
				pm.setRegion(&msg)
			case SetRuneDesc:
				log.Printf("set rune desc '%v' for channel: %v\n", msg.runeDesc, pm.ChannelID)
				pm.setRuneDesc(&msg)
			}

		case <-time.After(time.Second * 30):
			for _, p := range pm.FollowedPlayers {
				pm.handleMonitorPlayer(client(pm.region), p)
			}
			pm.reportGames()
		}
	}
}

func (pm *PlayerMonitor) setRuneDesc(msg *monitorMessage) {
	pm.RuneDesc = msg.runeDesc
	pm.save()
	message := newSuccessMessage(fmt.Sprintf("Rune description level be now '%v'.", strings.ToTitle(msg.runeDesc)), msg.timeStr, &msg.author)
	sendMessage(pm.ChannelID, message)
}

func (pm *PlayerMonitor) showGame(msg *monitorMessage) {
	summoner, err := client(pm.region).Summoner.SummonerByName(msg.summonerName)
	if err != nil || summoner == nil {
		DGBot.Discord.ChannelMessageSendComplex(pm.ChannelID, newErrorMessage(
			fmt.Sprintf("Unable t' find summoner %v from region %v", msg.summonerName, strings.ToTitle(pm.region)),
			msg.timeStr,
			&msg.author))
		return
	}

	cgi, err := client(pm.region).Spectator.ActiveGamesBySummoner(summoner.ID)
	if err != nil {
		DGBot.Discord.ChannelMessageSendComplex(pm.ChannelID, newErrorMessage(
			"Failed t' connect t' Riot API service",
			msg.timeStr,
			&msg.author))
		log.Printf("failed to fetch active games by summoner: %v", err)
		return
	}

	if cgi == nil {
		sendMessage(pm.ChannelID, newSuccessMessage(
			fmt.Sprintf("%v ain't in game", summoner.Name),
			msg.timeStr,
			&msg.author))
		return
	}

	pm.reportGame(cgi)
}

func (pm *PlayerMonitor) reportGame(cgi *riotapi.CurrentGameInfo) {
	bluePlayers := getPlayersOfTeam(teamBlue, cgi)
	blueReport, err := reportTeam(client(pm.region), "", bluePlayers, 0)
	if err != nil {
		log.Printf("failed to report team: %v", err)
		return
	}
	redPlayers := getPlayersOfTeam(teamRed, cgi)
	redReport, err := reportTeam(client(pm.region), "", redPlayers, len(bluePlayers))
	if err != nil {
		log.Printf("failed to report team: %v", err)
		return
	}

	blueReport.Color = blue
	redReport.Color = red
	redReport.Footer = &discordgo.MessageEmbedFooter{Text: "Enter players number to see rune information"}
	if len(blueReport.Fields) > 0 {
		sendMessage(pm.ChannelID, blueReport)
	}
	if len(redReport.Fields) > 0 {
		sendMessage(pm.ChannelID, redReport)
	}

	pm.games[cgi.GameID] = cgi
	pm.reportedGames[cgi.GameID] = true
	pm.lastGameID = cgi.GameID
}

func (pm *PlayerMonitor) setRegion(msg *monitorMessage) {
	pm.region = msg.region
	pm.FollowedPlayers = make(map[string]Player)
	pm.save()
	message := newSuccessMessage(fmt.Sprintf("Region be now '%v'. All followed players cleared.", strings.ToTitle(msg.region)), msg.timeStr, &msg.author)
	sendMessage(pm.ChannelID, message)
}

func (pm *PlayerMonitor) showRunes(playerIndex int, timeStr string, author *discordgo.User) {
	if len(pm.games) == 0 {
		return
	}
	if _, ok := pm.games[pm.lastGameID]; !ok {
		return
	}

	if len(pm.games[pm.lastGameID].Participants) <= playerIndex {
		DGBot.Discord.ChannelMessageSendComplex(pm.ChannelID, newErrorMessage("Invalid player number", timeStr, author))
		return
	}

	bluePlayer := getPlayersOfTeam(teamBlue, pm.games[pm.lastGameID])
	redPlayers := getPlayersOfTeam(teamRed, pm.games[pm.lastGameID])

	if len(bluePlayer) > 0 && playerIndex < len(bluePlayer) {
		me, err := pm.reportRuneForPlayer(bluePlayer[playerIndex], blue)
		if err != nil {
			log.Printf("Failed to report player data: %v\n", err)
			return

		}
		me.Footer = newFooter(author, timeStr)
		if _, err := DGBot.Discord.ChannelMessageSendComplex(pm.ChannelID, &discordgo.MessageSend{Embed: me}); err != nil {
			log.Printf("failed to send message: %v", err)
		}

	} else if len(redPlayers) > 0 {
		me, err := pm.reportRuneForPlayer(redPlayers[playerIndex-len(bluePlayer)], red)
		if err != nil {
			log.Printf("Failed to report player data: %v\n", err)
			return
		}
		me.Footer = newFooter(author, timeStr)
		if _, err := DGBot.Discord.ChannelMessageSendComplex(pm.ChannelID, &discordgo.MessageSend{Embed: me}); err != nil {
			log.Printf("failed to send message: %v", err)
		}
	}
}

func (pm *PlayerMonitor) removePlayer(msg *monitorMessage) {
	var removedSummoners []*discordgo.MessageEmbedField
	for _, p := range pm.FollowedPlayers {
		if strings.ToLower(p.Name) == strings.ToLower(msg.summonerName) {
			delete(pm.FollowedPlayers, strconv.Itoa(p.ID))
			removedSummoners = append(removedSummoners, &discordgo.MessageEmbedField{Name: p.Name, Value: p.Rank})
			break
		}
	}
	if len(removedSummoners) == 0 {
		DGBot.Discord.ChannelMessageSendComplex(pm.ChannelID, newErrorMessage(fmt.Sprintf("Unable t' find summoner %v", msg.summonerName), msg.timeStr, &msg.author))
		return
	}
	DGBot.Discord.ChannelMessageSendComplex(pm.ChannelID, &discordgo.MessageSend{
		Embed: newAddedSummonersMessage("Stopped followin'", msg.timeStr, &msg.author, removedSummoners),
	})
	pm.save()
}

func (pm *PlayerMonitor) listPlayers(timeStr string, author *discordgo.User) {
	var mefs []*discordgo.MessageEmbedField
	for _, p := range pm.FollowedPlayers {
		mefs = append(mefs, &discordgo.MessageEmbedField{Name: p.Name, Value: p.Rank})
	}
	msg := discordgo.MessageSend{
		Embed: &discordgo.MessageEmbed{
			Title:  "Followed summoners",
			Color:  green,
			Fields: mefs,
			Footer: newFooter(author, timeStr),
		},
	}
	DGBot.Discord.ChannelMessageSendComplex(pm.ChannelID, &msg)
}

func (pm *PlayerMonitor) toggleRunes(timeStr string, author *discordgo.User) {
	pm.ReportRunes = !pm.ReportRunes
	pm.save()
	reportRunes := "**-OFF-**"
	if pm.ReportRunes {
		reportRunes = "**-ON-**"
	}
	msg := discordgo.MessageSend{
		Embed: &discordgo.MessageEmbed{
			Title:  "Automatic runes reporting is now " + reportRunes,
			Color:  green,
			Footer: newFooter(author, timeStr),
		},
	}
	DGBot.Discord.ChannelMessageSendComplex(pm.ChannelID, &msg)
}

func (pm *PlayerMonitor) handleMonitorPlayer(c *riotapi.Client, p Player) {
	cgi, err := c.Spectator.ActiveGamesBySummoner(p.ID)
	if err != nil {
		log.Printf("failed to fetch active games by summoner: %v", err)
		return
	}
	if cgi == nil {
		if p.CurrentGameID > 0 {
			pm.endGame(pm.games[p.CurrentGameID])
			p.CurrentGameID = 0
		}
		return
	}

	if p.CurrentGameID > 0 {
		return
	}

	p.CurrentGameID = cgi.GameID
	pm.games[cgi.GameID] = cgi
	pm.lastGameID = cgi.GameID

}

func (pm *PlayerMonitor) endGame(g *riotapi.CurrentGameInfo) {
	if g == nil {
		return
	}
	log.Println("Game ended")
	delete(pm.games, g.GameID)
	delete(pm.reportedGames, g.GameID)
}

func (pm *PlayerMonitor) reportGames() {
	for _, game := range pm.games {
		if reported, ok := pm.reportedGames[game.GameID]; reported && ok {
			continue
		}
		pm.reportGame(game)
		if pm.ReportRunes {
			pm.reportGameRunes(game)
		}
	}
}

func (pm *PlayerMonitor) reportGameRunes(cgi *riotapi.CurrentGameInfo) {
	bluePlayers := getPlayersOfTeam(teamBlue, pm.games[pm.lastGameID])
	redPlayers := getPlayersOfTeam(teamRed, pm.games[pm.lastGameID])

	blueRunes, err := pm.reportRunes(bluePlayers, blue)
	if err != nil {
		log.Printf("failed to report runes: %v", err)
		return
	}
	for _, msg := range blueRunes {
		sendMessage(pm.ChannelID, msg)
	}
	redRunes, err := pm.reportRunes(redPlayers, red)
	if err != nil {
		log.Printf("failed to report runes: %v", err)
	}
	for _, msg := range redRunes {
		sendMessage(pm.ChannelID, msg)
	}
}

func (pm *PlayerMonitor) findGoodGuysTeamID(cgi *riotapi.CurrentGameInfo) int {
	for _, cgp := range cgi.Participants {
		if !pm.isPlayerNPC(&cgp) {
			return cgp.TeamID
		}
	}
	return teamNone
}

func (pm *PlayerMonitor) isPlayerNPC(cgp *riotapi.CurrentGameParticipant) bool {
	for _, p := range pm.FollowedPlayers {
		if p.ID == cgp.SummonerID {
			return false
		}
	}
	return true
}

func getOpposingTeam(teamID int) int {
	if teamID == teamBlue {
		return teamRed
	}
	return teamBlue
}

func reportTeam(c *riotapi.Client, title string, cgp []riotapi.CurrentGameParticipant, playerOffset int) (*discordgo.MessageEmbed, error) {
	fields := make([]*discordgo.MessageEmbedField, 0)
	for i, gp := range cgp {
		mef, err := npcMessageEmbedField(c, &gp, i+playerOffset)
		if err != nil {
			return nil, err
		}
		fields = append(fields, mef)
	}
	em := newEmbedMessage(title)
	em.Fields = fields
	return em, nil
}

func (pm *PlayerMonitor) reportRunes(cgp []riotapi.CurrentGameParticipant, color int) ([]*discordgo.MessageEmbed, error) {
	var mex []*discordgo.MessageEmbed
	for _, gp := range cgp {
		me, err := pm.reportRuneForPlayer(gp, color)
		if err != nil {
			return nil, err
		}
		mex = append(mex, me)
	}
	return mex, nil
}

func (pm *PlayerMonitor) reportRuneForPlayer(gp riotapi.CurrentGameParticipant, color int) (*discordgo.MessageEmbed, error) {
	champions, err := client(pm.region).StaticData.Champions()
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch champions data: %v\n", err)
	}

	rr, err := DDC.GetRunesReforged()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch runes: %v\n", err)
	}
	perkStyles := rr.PerkStyles
	runes := rr.AllRunes()

	var perkFields []*discordgo.MessageEmbedField
	for _, perkID := range gp.Perks.PerkIDs {
		perkEmo := DGBot.GetEmojiStr(strconv.Itoa(perkID))
		runeDesc := runes[perkID].ShortDesc
		if pm.RuneDesc == RuneLong {
			runeDesc = runes[perkID].LongDesc
		}

		field := discordgo.MessageEmbedField{
			Name:  fmt.Sprintf("%s %s", perkEmo, runes[perkID].Name),
			Value: runeDesc}
		perkFields = append(perkFields, &field)
	}

	url := "http://ddragon.leagueoflegends.com/cdn/img/champion/tiles/%s_0.jpg"
	thumbnailURL := fmt.Sprintf(url, strings.ToLower(sanitizeChampionName(champions.Data[gp.ChampionID].Name)))
	pStyleEmo := DGBot.GetEmojiStr(strconv.Itoa(gp.Perks.PerkStyle))
	pSubStyleEmo := DGBot.GetEmojiStr(strconv.Itoa(gp.Perks.PerkSubStyle))

	msg := discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    gp.SummonerName + " - " + champions.Data[gp.ChampionID].Name,
			IconURL: DDC.GetProfileIconURL(gp.ProfileIconID),
		},
		Description: fmt.Sprintf("**%v %v - %v %v**",
			pStyleEmo,
			perkStyles[gp.Perks.PerkStyle].Name,
			perkStyles[gp.Perks.PerkSubStyle].Name,
			pSubStyleEmo),
		Fields:    perkFields,
		Color:     color,
		Thumbnail: &discordgo.MessageEmbedThumbnail{URL: thumbnailURL},
	}
	return &msg, nil
}

func npcMessageEmbedField(c *riotapi.Client, cgp *riotapi.CurrentGameParticipant, playerNumber int) (*discordgo.MessageEmbedField, error) {
	champions, err := c.StaticData.Champions()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch champion data: %v", err)
	}
	rank, err := findPlayerRank(c, cgp.SummonerID)
	if err != nil {
		return nil, err
	}
	name := sanitizeChampionName(champions.Data[cgp.ChampionID].Name)
	champEmo := DGBot.GetEmojiStr(name)
	return &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("%d. %s %s", playerNumber, cgp.SummonerName, champEmo),
		Value:  rank,
		Inline: true,
	}, nil
}

func findPlayerRank(c *riotapi.Client, summonerID int) (string, error) {
	si, err := c.Summoner.SummonerByID(summonerID)
	if err != nil {
		return "", fmt.Errorf("failed to fetch summoner by id: %v", err)
	}
	recentMatches, err := c.Match.RecentMatchesByAccountID(si.AccountID)
	if err != nil {
		return "", fmt.Errorf("failed to fetch recent matches: %v", err)
	}

	if recentMatches != nil && len(recentMatches.Matches) > 0 {
		match, err := c.Match.MatchByID(recentMatches.Matches[0].GameID)
		if err != nil {
			return "", fmt.Errorf("failed to fetch match by id: %v", err)
		}
		var participantID int
		for _, ident := range match.ParticipantIdentities {
			if ident.Player.AccountID == si.AccountID {
				participantID = ident.ParticipantID
				break
			}
		}
		for _, p := range match.Participants {
			if p.ParticipantID == participantID {
				return p.HighestAchievedSeasonTier, nil
			}
		}
	}
	return "Not found", nil
}

func getPlayersOfTeam(teamID int, cgi *riotapi.CurrentGameInfo) []riotapi.CurrentGameParticipant {
	var team []riotapi.CurrentGameParticipant
	for _, cgp := range cgi.Participants {
		if cgp.TeamID == teamID {
			team = append(team, cgp)
		}
	}
	return team
}

func sanitizeChampionName(name string) string {
	return strings.ToLower(strings.Replace(strings.Replace(strings.Replace(name, "'", "", -1), " ", "", -1), ".", "", -1))
}
