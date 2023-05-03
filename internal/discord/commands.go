package discord

import (
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
	internalcommands "github.com/vaultbotx/vaultbot-lite/internal/commands"
	"github.com/vaultbotx/vaultbot-lite/internal/types"
)

func addTrack(s *discordgo.Session, i *discordgo.InteractionCreate) {
	trackId := i.ApplicationCommandData().Options[0].StringValue()

	meta := types.GetFieldsFromInteraction(i)
	track, err := internalcommands.AddTrack(context.Background(), trackId, meta)
	if err != nil {
		if err == types.ErrInvalidTrackId {
			err2 := respond(s, i, "Invalid track ID")
			if err2 != nil {
				log.WithFields(meta).Error(err2)
			}
		} else if err == types.ErrTrackAlreadyInPlaylist {
			err2 := respond(s, i, "Track is already in the playlist")
			if err2 != nil {
				log.WithFields(meta).Error(err2)
			}
		} else if err == types.ErrNoTrackExists {
			err2 := respond(s, i, "Track does not exist")
			if err2 != nil {
				log.WithFields(meta).Error(err2)
			}
		} else if err == types.ErrCouldNotAddToPlaylist ||
			err == types.ErrCouldNotAddToDatabase ||
			err == types.ErrCouldNotRemoveFromPlaylist {
			err2 := respond(s, i, "Could not add track to playlist. Please try again later")
			if err2 != nil {
				log.WithFields(meta).Error(err2)
			}
		}

		log.WithFields(meta).Error(err)
		return
	}

	trackDetails := fmt.Sprintf("%s by %s", track.Name, track.Artists[0].Name)
	err = respond(s, i, fmt.Sprintf("Added %s to the playlist", trackDetails))
	if err != nil {
		log.WithFields(meta).Error(err)
		return
	}
}

func respond(s *discordgo.Session, i *discordgo.InteractionCreate, response string) error {
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
					Name:        "trackId",
					Description: "The Spotify track ID, URI, or URL",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
				},
			},
		},
	}
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"add-track": addTrack,
	}
)
