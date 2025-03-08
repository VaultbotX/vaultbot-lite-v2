package blacklist

import (
	"github.com/bwmarrin/discordgo"
	"github.com/vaultbotx/vaultbot-lite/internal/discord/helpers"
)

var (
	minLength      = 1
	maxLengthUrl   = 256
	maxLengthGenre = 64
)

var BlacklistCommand = &discordgo.ApplicationCommand{
	Name:        "blacklist",
	Description: "Blacklist a track, artist, or genre",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Name:        "track",
			Description: "Blacklist a track (ID, URI, or URL)",
			Type:        discordgo.ApplicationCommandOptionString,
			Required:    false,
			MinLength:   &minLength,
			MaxLength:   maxLengthUrl,
		},
		{
			Name:        "artist",
			Description: "Blacklist an artist (ID, URI, or URL)",
			Type:        discordgo.ApplicationCommandOptionString,
			Required:    false,
			MinLength:   &minLength,
			MaxLength:   maxLengthUrl,
		},
		{
			Name:        "genre",
			Description: "Blacklist a genre",
			Type:        discordgo.ApplicationCommandOptionString,
			Required:    false,
			MinLength:   &minLength,
			MaxLength:   maxLengthGenre,
		},
	},
	DefaultMemberPermissions: &helpers.AdminPermission,
}

var UnblacklistCommand = &discordgo.ApplicationCommand{
	Name:        "unblacklist",
	Description: "Unblacklist a track, artist, or genre",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Name:        "track",
			Description: "Unblacklist a track (ID, URI, or URL)",
			Type:        discordgo.ApplicationCommandOptionString,
			Required:    false,
			MinLength:   &minLength,
			MaxLength:   maxLengthUrl,
		},
		{
			Name:        "artist",
			Description: "Unblacklist an artist (ID, URI, or URL)",
			Type:        discordgo.ApplicationCommandOptionString,
			Required:    false,
			MinLength:   &minLength,
			MaxLength:   maxLengthUrl,
		},
		{
			Name:        "genre",
			Description: "Unblacklist a genre",
			Type:        discordgo.ApplicationCommandOptionString,
			Required:    false,
			MinLength:   &minLength,
			MaxLength:   maxLengthGenre,
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
