package cron

import (
	"errors"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
	"github.com/vaultbotx/vaultbot-lite/internal/discord/helpers"
	"github.com/vaultbotx/vaultbot-lite/internal/utils"
)

var RefreshPlaylistCommand = &discordgo.ApplicationCommand{
	Name:        "refresh-playlist",
	Description: "Refresh a the selected playlist (outside of the normal schedule)",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Name:        "playlist",
			Description: "The playlist to refresh",
			Type:        discordgo.ApplicationCommandOptionString,
			Required:    true,
			Choices: []*discordgo.ApplicationCommandOptionChoice{
				{
					Name:  "Genre Playlist",
					Value: "genre",
				},
				{
					Name:  "High Scores Playlist",
					Value: "high_scores",
				},
			},
		},
	},
	DefaultMemberPermissions: &helpers.AdminPermission,
}

func RefreshPlaylistCommandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	selectedOption := i.ApplicationCommandData().Options[0]
	meta := utils.GetFieldsFromInteraction(i)

	log.WithFields(meta).Infof("Received refresh-playlist command with option: %s", selectedOption.StringValue())
	err := helpers.EnsureAdministratorRoleForUser(s, i)
	if err != nil {
		helpers.HandleUserPermissionError(s, i, err, meta)
		return
	}

	var refreshErr error
	switch selectedOption.StringValue() {
	case "genre":
		refreshErr = runPopulateGenrePlaylist()
	case "high_scores":
		refreshErr = runPopulateHighScoresPlaylist()
	default:
		refreshErr = errors.New("invalid playlist option selected: " + selectedOption.StringValue())
		return
	}

	if refreshErr != nil {
		log.WithFields(meta).Errorf("Error refreshing playlist: %s", err)
		err2 := helpers.RespondDelayed(s, i, "There was an error refreshing the selected playlist. Please try again later :(")
		if err2 != nil {
			log.WithFields(meta).Errorf("Error responding to user: %s", err2)
		}
		return
	}

	err = helpers.RespondDelayed(s, i, "Successfully refreshed the playlist!")
	if err != nil {
		log.WithFields(meta).Errorf("Error responding to user: %s", err)
		return
	}
}
