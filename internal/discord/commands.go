package discord

import (
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
	internalcommands "github.com/vaultbotx/vaultbot-lite/internal/commands"
	"github.com/vaultbotx/vaultbot-lite/internal/types"
	"github.com/vaultbotx/vaultbot-lite/internal/utils"
	"time"
)

func addTrack(s *discordgo.Session, i *discordgo.InteractionCreate) {
	trackId := i.ApplicationCommandData().Options[0].StringValue()

	meta := utils.GetFieldsFromInteraction(i)
	userFields := utils.GetUserFieldsFromInteraction(i)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	track, err := internalcommands.AddTrack(ctx, trackId, userFields, meta)
	cancel()

	if err != nil {
		switch err {
		case types.ErrInvalidTrackId:
			err2 := respond(s, i, "I can't recognize that track ID!")
			if err2 != nil {
				log.WithFields(meta).Error(err2)
			}
			break
		case types.ErrTrackAlreadyInPlaylist:
			err2 := respond(s, i, "Track is already in the playlist!")
			if err2 != nil {
				log.WithFields(meta).Error(err2)
			}
			break
		case types.ErrTrackTooLong:
			err2 := respond(s, i, "That track is too long!")
			if err2 != nil {
				log.WithFields(meta).Error(err2)
			}
		case types.ErrNoTrackExists:
			err2 := respond(s, i, "That track does not exist!")
			if err2 != nil {
				log.WithFields(meta).Error(err2)
			}
			break
		case types.ErrCouldNotAddToPlaylist:
		case types.ErrCouldNotAddToDatabase:
		case types.ErrCouldNotRemoveFromPlaylist:
			err2 := respond(s, i, "Could not add track to playlist. Please try again later :(")
			if err2 != nil {
				log.WithFields(meta).Error(err2)
			}
			break
		}

		log.WithFields(meta).Error(err)
		respond(s, i, "An unexpected error occurred. Please try again later :(")
		return
	}

	trackDetails := fmt.Sprintf("%s by %s", track.Name, track.Artists[0].Name)
	err = respond(s, i, fmt.Sprintf("Added %s to the playlist!", trackDetails))
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
					Name:        "track-id",
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
