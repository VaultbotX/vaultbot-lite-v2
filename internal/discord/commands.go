package discord

import (
	"os"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
	"github.com/vaultbotx/vaultbot-lite/internal/blacklist"
	"github.com/vaultbotx/vaultbot-lite/internal/cron"
	"github.com/vaultbotx/vaultbot-lite/internal/discord/helpers"
	"github.com/vaultbotx/vaultbot-lite/internal/preferences"
	"github.com/vaultbotx/vaultbot-lite/internal/tracks"
	"github.com/vaultbotx/vaultbot-lite/internal/utils"
)

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "ping",
			Description: "Ping the bot",
		},
		{
			Name:        "get-playlist",
			Description: "Get a link to the playlist",
		},
		{
			Name:        "refresh-genre-playlist",
			Description: "Refresh the genre playlist with a new genre (overrides daily schedule)",
		},
		preferences.Command,
		tracks.Command,
		blacklist.BlacklistCommand,
		blacklist.UnblacklistCommand,
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"ping": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			content := "Pong!"
			version, exists := os.LookupEnv("APP_VERSION")
			if exists {
				content += " (app version: " + version + ")"
			}

			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: content,
				},
			})
		},
		"get-playlist": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			playlistId := os.Getenv("SPOTIFY_PLAYLIST_ID")
			playlistUrl := "https://open.spotify.com/playlist/" + playlistId

			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Here is a link to the playlist: " + playlistUrl,
				},
			})
		},
		"refresh-genre-playlist": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			meta := utils.GetFieldsFromInteraction(i)

			err := helpers.RespondImmediately(s, i, "Processing your request...")
			if err != nil {
				log.WithFields(meta).Errorf("Error responding to user: %s", err)
				return
			}

			err = helpers.EnsureAdministratorRoleForUser(s, i)
			if err != nil {
				helpers.HandleUserPermissionError(s, i, err, meta)
				return
			}

			err = cron.RunPopulateGenrePlaylist()
			if err != nil {
				log.WithFields(meta).Errorf("Error refreshing genre playlist: %s", err)
				err2 := helpers.RespondDelayed(s, i, "There was an error refreshing the genre playlist. Please try again later :(")
				if err2 != nil {
					log.WithFields(meta).Errorf("Error responding to user: %s", err2)
				}
				return
			}

			err = helpers.RespondDelayed(s, i, "Successfully refreshed the genre playlist!")
			if err != nil {
				log.WithFields(meta).Errorf("Error responding to user: %s", err)
				return
			}
		},
		"add-track":        tracks.AddTrackCommandHandler,
		"edit-preferences": preferences.EditPreferencesCommandHandler,
		"blacklist":        blacklist.BlacklistCommandHandler,
		"unblacklist":      blacklist.UnblacklistCommandHandler,
	}
)
