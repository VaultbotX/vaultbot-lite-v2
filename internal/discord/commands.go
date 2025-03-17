package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/vaultbotx/vaultbot-lite/internal/blacklist"
	"github.com/vaultbotx/vaultbot-lite/internal/preferences"
	"github.com/vaultbotx/vaultbot-lite/internal/tracks"
	"os"
)

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "ping",
			Description: "Ping the bot",
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
				content += "(app version: " + version + ")"
			}

			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: content,
				},
			})
		},
		"add-track":        tracks.AddTrackCommandHandler,
		"edit-preferences": preferences.EditPreferencesCommandHandler,
		"blacklist":        blacklist.BlacklistCommandHandler,
		"unblacklist":      blacklist.UnblacklistCommandHandler,
	}
)
