package main

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/tahkapaa/gangplankbot/riotapi"
)

// Server ids for emojis here
var emojiGuilds = []string{
	"server id 1",
}

// Bot is a discordBot
type Bot struct {
	Discord  *discordgo.Session
	Channels map[string]*Channel
	db       DB
	Emojis   map[string]*discordgo.Emoji
}

// Channel is a channel with followed players
type Channel struct {
	ChannelID string
	monitor   *PlayerMonitor
	region    string
}

// Player is a lol player
type Player struct {
	Name          string
	ID            int
	CurrentGameID int
	Rank          string
}

func newBot(botToken string, db DB) (*Bot, error) {
	dg, err := discordgo.New("Bot " + botToken)
	if err != nil {
		return nil, err
	}
	bot := Bot{
		Discord:  dg,
		Channels: make(map[string]*Channel),
		db:       db,
		Emojis:   make(map[string]*discordgo.Emoji),
	}

	for _, guild := range emojiGuilds {
		if err := bot.initializeEmojis(guild); err != nil {
			return nil, err
		}
	}

	bot.readChannelsFromDB()
	bot.AddMessageHandler()

	if err := dg.Open(); err != nil {
		return nil, err
	}

	return &bot, nil
}

func (b *Bot) initializeEmojis(g string) error {
	guild, err := b.Discord.Guild(g)
	if err != nil {
		return fmt.Errorf("failed to get guild: %v", err)
	}

	for _, emo := range guild.Emojis {
		b.Emojis[emo.Name] = emo
	}
	return nil
}

// GetEmojiStr returns Emoji as string that can be sent to discord, or "" if emoji is not found
func (b *Bot) GetEmojiStr(id string) string {
	emoji, ok := DGBot.Emojis[id]
	if !ok {
		log.Printf("failed to map id '%v' to emoji", id)
		return ""
	}
	return fmt.Sprintf("<:%s>", emoji.APIName())
}

func (b *Bot) readChannelsFromDB() {
	channels, err := b.db.Get()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize channels from db: %v", err))
	}
	for _, channel := range channels {
		b.addChannel(channel.ID, channel.Region, channel.RuneDesc, channel.ReportRunes, channel.Summoners)
	}
}

func (b *Bot) addChannel(ID, region, runeDesc string, reportRunes bool, players map[string]Player) {
	if players == nil {
		players = make(map[string]Player)
	}
	pm := PlayerMonitor{
		FollowedPlayers: players,
		games:           make(map[int]*riotapi.CurrentGameInfo),
		reportedGames:   make(map[int]bool),
		messageChan:     make(chan monitorMessage, 1),
		ChannelID:       ID,
		db:              b.db,
		ReportRunes:     reportRunes,
		region:          region,
		RuneDesc:        runeDesc,
	}

	go pm.monitorPlayers()

	b.Channels[ID] = &Channel{
		ChannelID: ID,
		monitor:   &pm,
		region:    region,
	}
}

// AddMessageHandler adds CreateMessage handler to the bot
func (b *Bot) AddMessageHandler() {
	b.Discord.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		timeStr := time.Now().Format(time.ANSIC)
		_, ok := b.Channels[m.ChannelID]
		if !ok {
			b.addChannel(m.ChannelID, DefaultRegion, RuneShort, false, nil)
		}

		// Ignore all messages created by the bot itself
		// This isn't required in this specific example but it's a good practice.
		if m.Author.ID == s.State.User.ID {
			return
		}

		// If the message is "ping" reply with "Pong!"
		if m.Content == "ping" {
			s.ChannelMessageSend(m.ChannelID, "Yarrr!")
		}

		// If the message is "pong" reply with "Ping!"
		if m.Content == "pong" {
			s.ChannelMessageSend(m.ChannelID, "Yarrr!")
		}

		if strings.Contains(strings.ToLower(m.Content), "yarr") {
			s.ChannelMessageSend(m.ChannelID, "Yarrr!")
		}

		if strings.Contains(strings.ToLower(m.Content), "kapteeni") || strings.Contains(strings.ToLower(m.Content), "kapu") {
			s.ChannelMessageSend(m.ChannelID, "Yarrr!")
		}

		if m.Content == "?help" {
			b.handleHelp(m.ChannelID, timeStr, m.Author, s)
			return
		}

		// Start following players
		if strings.HasPrefix(m.Content, "?add") {
			b.Channels[m.ChannelID].handleStartFollowing(m.Content, timeStr, m.Author, s)
			return
		}

		// Stop following players
		if strings.HasPrefix(m.Content, "?remove") {
			b.Channels[m.ChannelID].handleStopFollowing(m.ChannelID, m.Content, timeStr, m.Author)
			return
		}

		// Report current game for player
		if strings.HasPrefix(m.Content, "?game") {
			b.Channels[m.ChannelID].handleShowGame(m.ChannelID, m.Content, timeStr, m.Author)
			return
		}

		// list followed players
		if m.Content == "?list" {
			b.Channels[m.ChannelID].handleListFollowedPlayers(m.ChannelID, timeStr, m.Author)
			return
		}

		if m.Content == "?joke" {
			s.ChannelMessageSend(m.ChannelID, jokes[rand.Intn(len(jokes))])
			return
		}

		// Toggle automatic runes reporting (spam alert)
		if m.Content == "?runes" {
			b.Channels[m.ChannelID].handleToggleRunes(timeStr, m.Author)
			return
		}

		// Set region for channel
		if strings.HasPrefix(m.Content, "?region") {
			b.Channels[m.ChannelID].handleSetRegion(m.Content, timeStr, m.Author, s)
			return
		}

		if strings.HasPrefix(m.Content, "?runedesc") {
			b.Channels[m.ChannelID].handleRuneDesc(m.ChannelID, m.Content, timeStr, m.Author)
			return
		}

		if num, err := strconv.Atoi(m.Content); err == nil {
			if num >= 0 && num <= 9 {
				b.Channels[m.ChannelID].handleShowPlayerRunes(num, timeStr, m.Author)
			}
		}
	})
}

func (b *Bot) handleHelp(channelID, timeStr string, author *discordgo.User, s *discordgo.Session) {
	reportRunes := "**-OFF-**"
	if b.Channels[channelID].monitor.ReportRunes {
		reportRunes = "**-ON-**"
	}
	msg := discordgo.MessageSend{
		Embed: &discordgo.MessageEmbed{
			Title: "Available commands",
			Color: green,
			Fields: []*discordgo.MessageEmbedField{
				{Name: "?add [name]", Value: "Summoner name t' be followed"},
				{Name: "?remove [name]", Value: "Summoner name that should nah be followed"},
				{Name: "?list", Value: "List o' summoners that are bein' followed"},
				{Name: "?game [name]", Value: "Report current game for summoner"},
				{Name: "?runes", Value: "Toggle automatic runes reportin' | " + reportRunes},
				{Name: "?runedesc [short, long]", Value: fmt.Sprintf("Set rune report description level | **%s**",
					strings.ToTitle(b.Channels[channelID].monitor.RuneDesc))},
				{Name: "?region [name]", Value: fmt.Sprintf("Set league o' legends region'. Clears all followed summoners! | **%s**", strings.ToTitle(b.Channels[channelID].region))},
				{Name: "?joke", Value: "Wants t' hear a joke?"},
			},
			Footer: newFooter(author, timeStr),
		},
	}
	s.ChannelMessageSendComplex(channelID, &msg)
}

func (c *Channel) handleStartFollowing(name, timeStr string, author *discordgo.User, s *discordgo.Session) {
	st, err := s.ChannelMessageSendComplex(c.ChannelID, newWorkingMessage())
	if err != nil {
		log.Println(err)
		return
	}
	name = c.removeCommandFromString(name, "?add", "The format is: ?add [name] - ?add player", timeStr, author)
	if name == "" {
		return
	}

	var addedSummoners []*discordgo.MessageEmbedField
	summoner, err := client(c.region).Summoner.SummonerByName(name)
	if err != nil || summoner == nil {
		s.ChannelMessageSendComplex(c.ChannelID, newErrorMessage(fmt.Sprintf("Unable t' find summoner: %v", name), timeStr, author))
		return
	}
	rank, err := findPlayerRank(client(c.region), summoner.ID)
	if err != nil {
		log.Println(err)
	}
	addedSummoners = append(addedSummoners, &discordgo.MessageEmbedField{Name: summoner.Name, Value: rank})

	c.monitor.messageChan <- monitorMessage{
		mtype:   AddPlayer,
		player:  Player{Name: summoner.Name, ID: summoner.ID, Rank: rank},
		author:  *author,
		timeStr: timeStr,
	}
	if len(addedSummoners) > 0 {
		s.ChannelMessageEditComplex(&discordgo.MessageEdit{
			Channel: c.ChannelID,
			ID:      st.ID,
			Embed:   newAddedSummonersMessage("Now followin'", timeStr, author, addedSummoners),
		})
	}
}

func (c *Channel) handleStopFollowing(channelID, command, timeStr string, author *discordgo.User) {
	name := c.removeCommandFromString(command, "?remove", "The format is: ?remove [name] - ?remove player", timeStr, author)
	if name == "" {
		return
	}

	c.monitor.messageChan <- monitorMessage{
		mtype:        RemovePlayer,
		summonerName: name,
		author:       *author,
		timeStr:      timeStr,
	}
}

func (c *Channel) handleShowGame(channelID, command, timeStr string, author *discordgo.User) {
	name := c.removeCommandFromString(command, "?game", "The format is: ?game [name] - ?game player", timeStr, author)
	if name == "" {
		return
	}

	c.monitor.messageChan <- monitorMessage{
		mtype:        ShowGame,
		summonerName: name,
		author:       *author,
		timeStr:      timeStr,
	}
}

func (c *Channel) handleListFollowedPlayers(channelID, timeStr string, author *discordgo.User) {
	c.monitor.messageChan <- monitorMessage{
		mtype:   ListPlayers,
		author:  *author,
		timeStr: timeStr,
	}
}

func (c *Channel) handleShowPlayerRunes(playerIndex int, timeStr string, author *discordgo.User) {
	c.monitor.messageChan <- monitorMessage{
		mtype:       ShowRunes,
		author:      *author,
		timeStr:     timeStr,
		playerIndex: playerIndex,
	}
}

func (c *Channel) handleToggleRunes(timeStr string, author *discordgo.User) {
	c.monitor.messageChan <- monitorMessage{
		mtype:   ToggleRunesReporting,
		author:  *author,
		timeStr: timeStr,
	}
}

func (c *Channel) handleSetRegion(region, timeStr string, author *discordgo.User, s *discordgo.Session) {
	split := strings.Fields(region)
	if len(split) != 2 {
		s.ChannelMessageSendComplex(c.ChannelID, newErrorMessage("The syntax is: ?region [region]. For example: ?region euw", timeStr, author))
		return
	}

	r := strings.ToLower(split[1])

	if _, ok := riotapi.APIHosts[r]; !ok {
		var regions []string
		for k := range riotapi.APIHosts {
			regions = append(regions, k)
		}
		message := fmt.Sprintf("Invalid region '%v', available regions: %v", split[1], strings.ToUpper(strings.Join(regions, ", ")))
		s.ChannelMessageSendComplex(c.ChannelID, newErrorMessage(message, timeStr, author))
		return
	}

	c.region = r

	c.monitor.messageChan <- monitorMessage{
		mtype:   SetRegion,
		region:  r,
		author:  *author,
		timeStr: timeStr,
	}
}

func (c *Channel) handleRuneDesc(channelID, command, timeStr string, author *discordgo.User) {
	errorMsg := "The format is: ?runedesc [short, long] - ?runedesc short"
	desc := strings.ToLower(c.removeCommandFromString(command, "?runedesc", errorMsg, timeStr, author))
	if desc == "" {
		return
	}
	if desc != RuneShort && desc != RuneLong {
		DGBot.Discord.ChannelMessageSendComplex(c.ChannelID, newErrorMessage(errorMsg, timeStr, author))
		return
	}

	c.monitor.messageChan <- monitorMessage{
		mtype:    SetRuneDesc,
		runeDesc: desc,
		author:   *author,
		timeStr:  timeStr,
	}
}

func (c *Channel) handleChampion(champion, timeStr string, author *discordgo.User, s *discordgo.Session) {
	split := strings.Fields(champion)
	if len(split) != 2 {
		s.ChannelMessageSendComplex(c.ChannelID, newErrorMessage("The syntax is: ?champion [champion]. For example: ?champion Corki", timeStr, author))
		return
	}

	emoji, ok := DGBot.Emojis[split[1]]
	if !ok {
		log.Printf("failed to map champion '%v' to emoji", champion)
		return
	}

	sendMessage(c.ChannelID, newSuccessMessage(fmt.Sprintf("<:%s>", emoji.APIName()), timeStr, author))
}

func (c *Channel) removeCommandFromString(s, cmd, errorMsg, timeStr string, author *discordgo.User) string {
	command := strings.Replace(s, cmd+" ", "", -1)
	if len(command) == 0 {
		DGBot.Discord.ChannelMessageSendComplex(c.ChannelID, newErrorMessage(errorMsg, timeStr, author))
		return ""
	}
	return command
}
