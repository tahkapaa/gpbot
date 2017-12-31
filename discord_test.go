package main

import (
	"fmt"
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestSend(t *testing.T) {

	m := discordgo.MessageEmbed{Type: "testii", Title: "OMG! Kuka vetäs ässän?", Color: 13125190,
		// Fields: []*discordgo.MessageEmbedField{
		// 	&discordgo.MessageEmbedField{Name: "Embed", Value: "Value", Inline: false},
		// 	&discordgo.MessageEmbedField{Name: "Embed2", Value: "Value2", Inline: true},
		// 	&discordgo.MessageEmbedField{Name: "Embed3", Value: "Value3", Inline: true},
		// },
		Footer: &discordgo.MessageEmbedFooter{Text: "Cpt. Selviö", IconURL: avatarURL},
	}

	sendMessages([]*discordgo.MessageEmbed{&m})
}

func TestDiscord(t *testing.T) {
	dg, err := discordgo.New("Bot Mzg3MzQwODQ3NjUxMjI1NjAw.DQdjMQ.ER4bcXCinp9wnhODkXsmPhDZV8I")
	if err != nil {
		t.Fatal(err.Error())
	}

	st, err := dg.GuildChannels("386887549408247826")
	if err != nil {
		t.Fatal("err: ", err.Error())
	}
	fmt.Println(st[1])

	msg, err := dg.ChannelMessageSend(st[1].ID, "Testi")
	if err != nil {
		t.Fatal("Send fail: ", err)
	}
	fmt.Println(msg)
}
