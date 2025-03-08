package blacklist

import (
	"github.com/bwmarrin/discordgo"
	"github.com/vaultbotx/vaultbot-lite/internal/discord/helpers"
)

var Command = &discordgo.ApplicationCommand{
	Name:        "blacklist",
	Description: "Blacklist a track, artist, or genre",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Name:        "type",
			Description: "The type of item to blacklist",
			Type:        discordgo.ApplicationCommandOptionString,
			Required:    true,
			Choices: []*discordgo.ApplicationCommandOptionChoice{
				{
					Name:  "track",
					Value: "track",
				},
				{
					Name:  "artist",
					Value: "artist",
				},
				{
					Name:  "genre",
					Value: "genre",
				},
			},
		},
	},
	DefaultMemberPermissions: &helpers.AdminPermission,
}

func BlacklistCommandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	blacklist(s, i, true)
}

func UnblacklistCommandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	blacklist(s, i, false)
}
