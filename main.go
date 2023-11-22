package main

import (
	// "flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

// Variables used for command line parameters
var (
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
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "you sent: " + url,
				},
			})
		},
	}
)

func init() {
	// flag.StringVar(&Token, "t", "", "Bot Token")
	// flag.Parse()

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

	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {

		name := i.ApplicationCommandData().Name

		if handlerFunc, ok := commandHandlers[name]; ok {
			handlerFunc(s, i)
		}
	})

	// Register the messageCreate func as a callback for MessageCreate events.
	// dg.AddHandler(messageCreate)

	// In this example, we only care about receiving message events.
	dg.Identify.Intents = discordgo.IntentsGuildMessages

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}
	fmt.Println("ws connection opened")

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

// // This function will be called (due to AddHandler above) every time a new
// // message is created on any channel that the authenticated bot has access to.
// func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

// 	// Ignore all messages created by the bot itself
// 	// This isn't required in this specific example but it's a good practice.
// 	if m.Author.ID == s.State.User.ID {
// 		return
// 	}
// 	// If the message is "ping" reply with "Pong!"
// 	if m.Content == "ping" {
// 		s.ChannelMessageSend(m.ChannelID, "Pong!")
// 	}

// 	// If the message is "pong" reply with "Ping!"
// 	if m.Content == "pong" {
// 		s.ChannelMessageSend(m.ChannelID, "Ping!")
// 	}
// }
