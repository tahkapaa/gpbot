package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/tahkapaa/gangplankbot/riotapi"
)

var (
	// Token contains the Discord Bot token
	Token string

	// RiotKey contains the Riot API key
	RiotKey string

	// DefaultRegion contains the region that should be used as default
	DefaultRegion string
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.StringVar(&RiotKey, "a", "", "Riot API key")
	flag.StringVar(&DefaultRegion, "r", "euw", "Default region to use")
	flag.Parse()
}

var clientMap sync.Map

// DDC is the riotapi DataDragon client
var DDC *riotapi.DDragonClient

// DGBot is the discord bot
var DGBot *Bot

func main() {
	rand.Seed(time.Now().UnixNano())

	createRiotClients(RiotKey, 50, 20)
	DDC = riotapi.NewDDragonClient()

	bot, err := newBot(Token, New("channels"))
	if err != nil {
		log.Fatalf("Unable to create bot: %v", err)
	}
	DGBot = bot

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	bot.Discord.Close()
}

func createRiotClients(apikey string, requestsPerMinute, burst int) {
	for k := range riotapi.APIHosts {
		c, err := riotapi.New(RiotKey, k, requestsPerMinute, burst)
		if err != nil {
			log.Fatalf("unable to initialize riot api: %v", err)
		}
		clientMap.Store(k, c)
	}
}

func client(region string) *riotapi.Client {
	if region == "" {
		region = DefaultRegion
	}
	c, ok := clientMap.Load(region)
	if !ok {
		log.Fatalf("region not found from clientmap: %v", region)
	}
	return c.(*riotapi.Client)
}
