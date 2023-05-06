package discord

import (
	"context"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	internalcommands "github.com/vaultbotx/vaultbot-lite/internal/commands"
	"io"
	"os"
	"os/signal"
	"time"
)

var (
	s           *discordgo.Session
	TestGuildId string
	BotToken    string
)

func Run() {
	log.SetFormatter(&log.JSONFormatter{})

	_, envPresent := os.LookupEnv("ENVIRONMENT")
	if !envPresent {
		log.SetLevel(log.DebugLevel)
		log.Info("Starting in development mode")

		err := godotenv.Load("dev.env")
		if err != nil {
			log.Fatal("Error loading dev.env file")
		}
	} else {
		log.SetLevel(log.InfoLevel)
		log.Info("Starting in production mode")
	}

	tokenPresent := true
	BotToken, tokenPresent = os.LookupEnv("DISCORD_TOKEN")
	if !tokenPresent {
		log.Fatal("Missing DISCORD_TOKEN environment variable")
	}

	guildId, guildIdPresent := os.LookupEnv("DISCORD_GUILD_ID")
	if !guildIdPresent {
		log.Debug("DISCORD_GUILD_ID environment variable missing, commands will be registered globally")
		TestGuildId = ""
	} else {
		TestGuildId = guildId
	}

	var err error
	s, err = discordgo.New("Bot " + BotToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}

	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Infof("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	err = internalcommands.CacheTracks(ctx)
	if err != nil {
		log.Fatalf("Cannot cache playlist tracks: %v", err)
	}
	RunPurge()
	cancel()

	err = s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	defer func(s *discordgo.Session) {
		err := s.Close()
		if err != nil {
			log.Fatalf("Cannot gracefully close the session: %v", err)
		}
	}(s)

	log.Info("Adding commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err2 := s.ApplicationCommandCreate(s.State.User.ID, TestGuildId, v)
		if err2 != nil {
			if restErr, ok := err2.(*discordgo.RESTError); ok {
				message := ""
				if restErr.Message != nil {
					message = restErr.Message.Message
				}

				bodyString := ""
				body, err3 := restErr.Request.GetBody()
				if err3 == nil {
					bodyBytes, err4 := io.ReadAll(body)
					if err4 == nil {
						bodyString = string(bodyBytes)
					}
				}

				log.WithFields(log.Fields{
					"command":         v.Name,
					"responseMessage": message,
					"requestBody":     bodyString,
				}).Fatalf("Request to create application command failed with response code %d",
					restErr.Response.StatusCode)
			}

			log.Fatalf("Cannot create '%v' command: %v", v.Name, err2)
		}
		registeredCommands[i] = cmd
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, os.Kill)
	<-stop

	for _, v := range registeredCommands {
		err := s.ApplicationCommandDelete(s.State.User.ID, TestGuildId, v.ID)
		if err != nil {
			log.Fatalf("Cannot delete '%v' command: %v", v.Name, err)
		}
	}

	log.Info("Gracefully shutting down.")
}
