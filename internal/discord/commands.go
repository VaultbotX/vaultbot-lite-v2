package discord

import (
	"github.com/bwmarrin/discordgo"
	discordcommands "github.com/vaultbotx/vaultbot-lite/internal/discord/commands"
)

func Respond(s *discordgo.Session, i *discordgo.InteractionCreate, response string) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: response,
		},
	})
}

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "add-track",
			Description: "Add a track to the playlist",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "track-id",
					Description: "The Spotify track ID, URI, or URL",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
				},
			},
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"add-track": discordcommands.AddTrack,
	}
)
