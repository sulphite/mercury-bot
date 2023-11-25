package main

import (
	// "flag"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/mmcdole/gofeed"
)

type Feed struct {
	Url        string
	Name       string
	Channel_id string
	Last_guid  string
}

// type Server struct {
// 	server_id string
// 	feeds     []Feed
// }

// Variables used for command line parameters
var (
	mu     sync.Mutex
	config []Feed
	// 	Token string
	// create command structure
	// every command needs a name and description!
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "test",
			Description: "Basic command",
		},
		{
			Name:        "sub",
			Description: "subscribe to a feed",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "url",
					Description: "The URL of the feed",
					Required:    true,
				},
			},
		},
	}
	// create command handlers
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"test": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Hello i'm mercury!",
				},
			})
		},
		"sub": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			url := i.ApplicationCommandData().Options[0].StringValue()
			fp := gofeed.NewParser()
			feed, err := fp.ParseURL(url)
			if err != nil {
				panic(err)
			}
			newFeed := Feed{
				Url:        url,
				Name:       feed.Title,
				Channel_id: i.ChannelID,
				Last_guid:  feed.Items[0].GUID,
			}
			mu.Lock()
			defer mu.Unlock()
			config = append(config, newFeed)
			embeds := make([]*discordgo.MessageEmbed, 1)
			embeds = append(embeds, createEmbed(feed.Items[0]))

			// response
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "You subscribed to " + feed.Title,
					Embeds:  embeds,
				},
			})
		},
	}
)

func init() {
	mu.Lock()
	defer mu.Unlock()
	data, e := os.ReadFile("bot_config.json")
	if e != nil {
		panic(e)
	}

	json.Unmarshal(data, &config)

}

func main() {

	// get token from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	Token := os.Getenv("TOKEN")
	AppID := os.Getenv("APP_ID")

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}
	fmt.Println("Discord session successfully created")

	fmt.Println("registering commands...")
	for _, command := range commands {
		dg.ApplicationCommandCreate(AppID, "", command)
	}
	fmt.Println("commands registered.")

	// add a single handler that will find the correct handler and run it
	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {

		name := i.ApplicationCommandData().Name

		if handlerFunc, ok := commandHandlers[name]; ok {
			handlerFunc(s, i)
		}
	})

	// In this example, we only care about receiving message events.
	dg.Identify.Intents = discordgo.IntentsGuildMessages

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}
	fmt.Println("ws connection opened")
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")

	done := make(chan bool)
	// check feeds regularly
	go runScheduler(dg, &config, done)

	// Wait here until CTRL-C or other term signal is received.
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// shut down other goroutine
	done <- true

	mu.Lock()
	configJson, _ := json.Marshal(config)
	mu.Unlock()
	err = writeFile("bot_config.json", configJson)
	if err != nil {
		log.Println("data was not saved successfully")
		log.Println(err)
	}

	// Cleanly close down the Discord session.
	dg.Close()
}

func writeFile(path string, data []byte) error {
	err := os.WriteFile(path, data, 0600)
	return err
}

func runScheduler(session *discordgo.Session, config *[]Feed, done chan bool) {
	fp := gofeed.NewParser()
	ticker := time.NewTicker(2 * time.Minute)
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			fmt.Println("checking feeds...")
			for i := range *config {
				feed := &(*config)[i]
				feeddata, err := fp.ParseURL(feed.Url)
				if err != nil {
					log.Println(err)
				}
				top := feeddata.Items[0]
				log.Print("top: ", top.GUID, "last: ", feed.Last_guid)
				if top.GUID != feed.Last_guid {
					session.ChannelMessageSendComplex(feed.Channel_id, &discordgo.MessageSend{
						Content: "New content!",
						Embeds:  []*discordgo.MessageEmbed{createEmbed(top)},
					})
					mu.Lock()
					feed.Last_guid = top.GUID
					mu.Unlock()
				}
			}
		}
	}
}

func createEmbed(feeditem *gofeed.Item) *discordgo.MessageEmbed {
	myembedptr := &discordgo.MessageEmbed{
		Title:       feeditem.Title,
		URL:         feeditem.Link,
		Type:        "link",
		Description: feeditem.Description,
	}
	log.Println(*myembedptr)
	return myembedptr
}
