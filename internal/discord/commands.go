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
			// TODO: Will likely want to make this a single command with subcommands
			Name:        "edit-preference-track-duration",
			Description: "Edit the track duration preference",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "track-duration",
					Description: "The track duration in minutes",
					Type:        discordgo.ApplicationCommandOptionInteger,
					Required:    true,
					MinValue:    &MinSongDuration,
					MaxValue:    MaxSongDuration,
				},
			},
			DefaultMemberPermissions: &AdminPermission,
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"add-track":                      discordcommands.AddTrack,
		"edit-preference-track-duration": discordcommands.EditPreferenceTrackDuration,
	}
)
