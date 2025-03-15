package discord

import (
	"context"
	"errors"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/vaultbotx/vaultbot-lite/internal/preferences"
	"github.com/vaultbotx/vaultbot-lite/internal/tracks"
	"io"
	"net/http"
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

	loadEnvVars()

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

	s.AddHandler(func(s *discordgo.Session, e *discordgo.Resumed) {
		log.Info("Resumed session")
	})

	// TODO: Attempt to use least privileged intents. The commented intents were not sufficient
	//  to retrieve a user's permissions in a guild.
	// Used for calculating guild member permissions
	//s.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMembers
	s.Identify.Intents = discordgo.IntentsAll

	startBackgroundTasks()

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

	registeredCommands := addDiscordCommands()

	ctx, cancel := context.WithCancel(context.Background())

	addHealthCheck(ctx)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, os.Kill)

	// Block until a signal is received
	<-stop
	log.Info("Received shutdown signal, exiting gracefully")
	cancel()

	// Teardown
	_, envPresent := os.LookupEnv("ENVIRONMENT")
	if envPresent {
		log.Info("Cleaning up registered commands")
		for _, v := range registeredCommands {
			err := s.ApplicationCommandDelete(s.State.User.ID, TestGuildId, v.ID)
			if err != nil {
				log.Fatalf("Cannot delete '%v' command: %v", v.Name, err)
			}
		}
	}

	log.Info("Gracefully shutting down.")
}

// addHealthCheck spins up a health check on /api/healthz
func addHealthCheck(ctx context.Context) {
	go func() {
		log.Info("Starting health check server")

		mux := http.NewServeMux()
		mux.HandleFunc("/api/healthz", func(w http.ResponseWriter, r *http.Request) {
			select {
			case <-ctx.Done():
				log.Warn("Health check request cancelled")
				http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
				return
			default:
				// continue processing the request
			}

			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("OK"))
			if err != nil {
				log.Errorf("Failed to write health check response: %v", err)
			}
		})

		err := http.ListenAndServe(":8080", mux)
		if err != nil {
			log.Fatalf("Failed to start health check server: %v", err)
		}
	}()
}

func addDiscordCommands() []*discordgo.ApplicationCommand {
	log.Info("Adding Discord commands")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err2 := s.ApplicationCommandCreate(s.State.User.ID, TestGuildId, v)
		if err2 != nil {
			var restErr *discordgo.RESTError
			if errors.As(err2, &restErr) {
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
		} else {
			log.Infof("Created '%v' command", cmd.Name)
		}
		registeredCommands[i] = cmd
	}
	log.Info("Finished adding Discord commands")
	return registeredCommands
}

func startBackgroundTasks() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	log.Debug("Caching tracks")
	err := tracks.CacheTracks(ctx)
	if err != nil {
		log.Fatalf("Cannot cache playlist tracks: %v", err)
	}
	log.Debug("Finished caching tracks")

	log.Debug("Checking default preferences")
	err = preferences.CheckDefaultPreferences(ctx)
	if err != nil {
		log.Fatalf("Cannot check default preferences: %v", err)
	}
	log.Debug("Finished checking default preferences")

	log.Debug("Starting purge tracks cron")
	RunPurge()
	log.Debug("Finished starting purge tracks cron")

	cancel()
}

func loadEnvVars() {
	_, envPresent := os.LookupEnv("ENVIRONMENT")
	if envPresent {
		log.SetLevel(log.DebugLevel)
		log.Info("Starting in development mode")

		err := godotenv.Load(".env")
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	} else {
		log.SetLevel(log.DebugLevel)
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
}
