package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/bwmarrin/discordgo"
)

const (
	discordHook = "https://discordapp.com/api/webhooks/384835834815447052/K6amIwt30YVjWBJFivZIR8UIBB8Qh-mUGcleVUQ0oTSTt5BJuR0eXRKZ1xJyqEmEzscF"
	avatarURL   = "http://ddragon.leagueoflegends.com/cdn/img/champion/tiles/gangplank_0.jpg"
)

const (
	red    = 16750480
	blue   = 2926540
	green  = 5308359
	yellow = 16773522
)

func sendMessage(channelID string, msg *discordgo.MessageEmbed) {
	if _, err := DGBot.Discord.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{Embed: msg}); err != nil {
		log.Printf("failed to send message to discord: %v", err)
	}
}

func sendToDiscord(s string) {
	dm := discordgo.WebhookParams{
		Content:   fmt.Sprintf("%s", s),
		AvatarURL: avatarURL,
	}

	b, err := json.Marshal(dm)
	if err != nil {
		panic(err)
	}

	resp, err := http.Post(discordHook, "application/json", bytes.NewReader(b))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
}

func newWorkingMessage() *discordgo.MessageSend {
	return &discordgo.MessageSend{
		Embed: &discordgo.MessageEmbed{
			Title: "Runnin' on some errands...",
			Color: yellow,
		},
	}
}

func newEmbedMessage(title string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title: title,
		Color: blue,
	}
}

func newSuccessMessage(title, msgTime string, author *discordgo.User) *discordgo.MessageEmbed {
	return newMessage(title, msgTime, author, green)
}

func newErrorMessage(title, msgTime string, author *discordgo.User) *discordgo.MessageSend {
	return &discordgo.MessageSend{
		Embed: newMessage(title, msgTime, author, red),
	}
}

func newMessage(title, msgTime string, author *discordgo.User, color int) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:  title,
		Color:  color,
		Footer: newFooter(author, msgTime),
	}
}

func newFooter(author *discordgo.User, msgTime string) *discordgo.MessageEmbedFooter {
	return &discordgo.MessageEmbedFooter{
		Text:    fmt.Sprintf("Requested by: %s | %v", author.Username, msgTime),
		IconURL: author.AvatarURL("32"),
	}
}

func newAddedSummonersMessage(title, msgTime string, author *discordgo.User, fields []*discordgo.MessageEmbedField) *discordgo.MessageEmbed {
	msgEmbed := newSuccessMessage(title, msgTime, author)
	msgEmbed.Fields = fields
	return msgEmbed
}
