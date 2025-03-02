package tracks

import (
	"context"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
	"github.com/vaultbotx/vaultbot-lite/internal/discord/commands"
	"github.com/vaultbotx/vaultbot-lite/internal/persistence"
	"github.com/vaultbotx/vaultbot-lite/internal/persistence/postgres"
	"github.com/vaultbotx/vaultbot-lite/internal/types"
	"github.com/vaultbotx/vaultbot-lite/internal/utils"
	"time"
)

func AddTrackCommandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	trackId := i.ApplicationCommandData().Options[0].StringValue()

	meta := utils.GetFieldsFromInteraction(i)
	userFields := utils.GetUserFieldsFromInteraction(i)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)

	pgConn, err := postgres.NewPostgresConnection()
	if err != nil {
		log.WithFields(meta).Error(err)
		err2 := commands.Respond(s, i, "An unexpected error occurred. Please try again later :(")
		if err2 != nil {
			log.WithFields(meta).Error(err2)
			cancel()
			return
		}
		cancel()
		return
	}

	trackRepository := persistence.NewPostgresTrackRepository(pgConn)
	trackService := NewTrackService(trackRepository)
	track, err := AddTrack(trackService, trackId, userFields, ctx, meta)
	cancel()

	if err != nil {
		switch {
		case errors.Is(err, types.ErrInvalidTrackId):
			err2 := commands.Respond(s, i, "I can't recognize that track ID!")
			if err2 != nil {
				log.WithFields(meta).Error(err2)
			}
			break
		case errors.Is(err, types.ErrTrackAlreadyInPlaylist):
			err2 := commands.Respond(s, i, "Track is already in the playlist!")
			if err2 != nil {
				log.WithFields(meta).Error(err2)
			}
			break
		case errors.Is(err, types.ErrTrackTooLong):
			err2 := commands.Respond(s, i, "That track is too long!")
			if err2 != nil {
				log.WithFields(meta).Error(err2)
			}
		case errors.Is(err, types.ErrNoTrackExists):
			err2 := commands.Respond(s, i, "That track does not exist!")
			if err2 != nil {
				log.WithFields(meta).Error(err2)
			}
			break
		case errors.Is(err, types.ErrCouldNotAddToPlaylist):
		case errors.Is(err, types.ErrCouldNotAddToDatabase):
		case errors.Is(err, types.ErrCouldNotRemoveFromPlaylist):
			err2 := commands.Respond(s, i, "Could not add track to playlist. Please try again later :(")
			if err2 != nil {
				log.WithFields(meta).Error(err2)
			}
			break
		}

		log.WithFields(meta).Error(err)
		err2 := commands.Respond(s, i, "An unexpected error occurred. Please try again later :(")
		if err2 != nil {
			log.WithFields(meta).Error(err2)
			return
		}
		return
	}

	trackDetails := fmt.Sprintf("%s by %s", track.Name, track.Artists[0].Name)
	err = commands.Respond(s, i, fmt.Sprintf("Added %s to the playlist!", trackDetails))
	if err != nil {
		log.WithFields(meta).Error(err)
		return
	}
}
