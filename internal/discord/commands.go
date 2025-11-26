package discord

import (
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/vaultbotx/vaultbot-lite/internal/blacklist"
	"github.com/vaultbotx/vaultbot-lite/internal/preferences"
	"github.com/vaultbotx/vaultbot-lite/internal/tracks"
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
		"add-track":        tracks.AddTrackCommandHandler,
		"edit-preferences": preferences.EditPreferencesCommandHandler,
		"blacklist":        blacklist.BlacklistCommandHandler,
		"unblacklist":      blacklist.UnblacklistCommandHandler,
	}
)
