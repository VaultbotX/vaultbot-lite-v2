package discord

import (
	"github.com/bwmarrin/discordgo"
	discordcommands "github.com/vaultbotx/vaultbot-lite/internal/discord/commands"
)

var (
	// AdminPermission is the int64 representation of an admin permission.
	// Not aware of any constants in Discordgo to represent the permission int64s
	// https://discord-api-types.dev/api/discord-api-types-payloads/common#PermissionFlagsBits
	// https://github.com/discordjs/discord-api-types/blob/0e6b19d2bcfe6e9806d3d20125668b3464845517/payloads/common.ts#L26
	AdminPermission int64 = 8

	MinSongDuration float64 = 2
	MaxSongDuration float64 = 120

	MinPurgeFrequency float64 = 1
	MaxPurgeFrequency float64 = 7

	MinTrackAge float64 = 1
	MaxTrackAge float64 = 31
)

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
		{
			Name:        "edit-preferences",
			Description: "Edit preferences",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "max-track-duration",
					Description: "The track duration in minutes",
					Type:        discordgo.ApplicationCommandOptionInteger,
					Required:    false,
					MinValue:    &MinSongDuration,
					MaxValue:    MaxSongDuration,
				},
				{
					Name:        "purge-frequency",
					Description: "How often to purge the playlist (in days)",
					Type:        discordgo.ApplicationCommandOptionInteger,
					Required:    false,
					MinValue:    &MinPurgeFrequency,
					MaxValue:    MaxPurgeFrequency,
				},
				{
					Name:        "max-track-age",
					Description: "The maximum age of a track in days",
					Type:        discordgo.ApplicationCommandOptionInteger,
					Required:    false,
					MinValue:    &MinTrackAge,
					MaxValue:    MaxTrackAge,
				},
			},
			DefaultMemberPermissions: &AdminPermission,
		},
		{
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
			DefaultMemberPermissions: &AdminPermission,
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"add-track":        discordcommands.AddTrack,
		"edit-preferences": discordcommands.EditPreferences,
		"blacklist":        discordcommands.Blacklist,
	}
)
