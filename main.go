package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Variables used for command line parameters
var (
	CaseMap = cases.Title(language.AmericanEnglish)
	Token   string
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
	Filetype string
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
			ext := filepath.Ext(name)
			gophers = append(gophers, Gopher{
				strings.TrimSuffix(name, ext),
				name,
				strings.TrimPrefix(ext, "."),
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

// Handles the given input message, returning the response message and any open files.
func ParseCommand(m *discordgo.MessageCreate) (*discordgo.MessageSend, []*os.File, error) {
	// `!gopher [?name]`
	if b, err := regexp.MatchString(`^!gopher( [-A-Za-z]+)?$`, m.Content); err != nil {
		return nil, nil, fmt.Errorf("Failed to compile pattern!")
	} else if b {
		msg := strings.SplitN(m.Content, " ", 2)[1:]

		var gopher *Gopher
		var content_msg string
		if len(msg) == 0 {
			gopher, err = GetGopher("", true)
			if err != nil {
				return nil, nil, err
			}
			content_msg = "Random gopher go!"
		} else {
			gopher, err = GetGopher(msg[0], false)
			if err != nil {
				content_msg = fmt.Sprintf("I don't know this \"%s\" guy", msg[0])
				return &discordgo.MessageSend{Content: content_msg}, nil, nil
			}
			content_msg = fmt.Sprintf("Introducing Ser %s", CaseMap.String(gopher.Name))
		}
		file, err := os.Open(gopher.Path)
		if err != nil {
			return nil, nil, err
		}

		return &discordgo.MessageSend{
			Content: content_msg,
			File: &discordgo.File{
				Name:        gopher.Filename,
				ContentType: gopher.Filetype,
				Reader:      file,
			},
		}, []*os.File{file}, nil
	}

	// `!gophers`
	if b, err := regexp.MatchString(`^!gophers$`, m.Content); err != nil {
		return nil, nil, fmt.Errorf("Failed to compile pattern!")
	} else if b {
		gophers, err := ListGophers()
		if err != nil {
			return nil, nil, err
		}

		var output strings.Builder
		for _, gopher := range gophers {
			output.WriteString(gopher.Name + "\n")
		}

		return &discordgo.MessageSend{Content: output.String()}, nil, nil
	}
	return nil, nil, fmt.Errorf("Unknown command: \"%s\"", m.Content)
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	msg, files, err := ParseCommand(m)
	if err != nil {
		fmt.Println(err)
	}
	_, err = s.ChannelMessageSendComplex(m.ChannelID, msg)
	if err != nil {
		fmt.Println(err)
	}
	if files != nil {
		for _, file := range files {
			file.Close()
		}
	}
}
