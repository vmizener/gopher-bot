package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

// Variables used for command line parameters
var (
	Token string
)

const GopherPath = "./gophers/"

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
}

func main() {
	// Parse flags here (not `init`) to simplify testing.
	flag.Parse()

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	// In this example, we only care about receiving message events.
	dg.Identify.Intents = discordgo.IntentsGuildMessages

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

type Gopher struct {
	Name     string
	Filename string
	Path     string
}

func ListGophers() ([]Gopher, error) {
	file, err := os.Open(GopherPath)
	if err != nil {
		return nil, err
	}
	names, err := file.Readdirnames(0)
	if err != nil {
		return nil, err
	}
	var gophers []Gopher
	for _, name := range names {
		if strings.HasSuffix(name, ".png") {
			gophers = append(gophers, Gopher{
				strings.TrimSuffix(name, ".png"),
				name,
				GopherPath + name,
			})
		}
	}
	return gophers, nil
}

func GetGopher(name string, random bool) (*Gopher, error) {
	gophers, err := ListGophers()
	if err != nil {
		return nil, err
	}

	if random {
		return &gophers[rand.Intn(len(gophers))], nil
	}
	for _, gopher := range gophers {
		if gopher.Name == name {
			return &gopher, nil
		}
	}
	return nil, fmt.Errorf("No gopher with name %s", name)
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content == "!gopher" {

		// Retrieve a random gopher
		gopher, err := GetGopher("", true)
		if err != nil {
			fmt.Println(err)
		}
		file, err := os.Open(gopher.Path)
		if err != nil {
			fmt.Println(err)
		}
		defer file.Close()

		_, err = s.ChannelFileSend(m.ChannelID, gopher.Filename, file)
		if err != nil {
			fmt.Println(err)
		}
	}
	if m.Content == "!gophers" {

		// Get a list of gophers
		gophers, err := ListGophers()
		if err != nil {
			fmt.Println(err)
		}

		var output strings.Builder
		for _, gopher := range gophers {
			output.WriteString(gopher.Name + "\n")
		}

		// Send a text message with the list of Gophers
		_, err = s.ChannelMessageSend(m.ChannelID, output.String())
		if err != nil {
			fmt.Println(err)
		}
	}
}
