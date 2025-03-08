package tracks

import (
	"context"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
	"github.com/vaultbotx/vaultbot-lite/internal/discord/helpers"
	"github.com/vaultbotx/vaultbot-lite/internal/domain"
	"github.com/vaultbotx/vaultbot-lite/internal/persistence"
	"github.com/vaultbotx/vaultbot-lite/internal/persistence/postgres"
	"github.com/vaultbotx/vaultbot-lite/internal/spotify"
	sp "github.com/vaultbotx/vaultbot-lite/internal/spotify/commands"
	"github.com/vaultbotx/vaultbot-lite/internal/utils"
	"time"
)

func AddTrackCommandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	trackId := i.ApplicationCommandData().Options[0].StringValue()

	meta := utils.GetFieldsFromInteraction(i)
	userFields := utils.GetUserFieldsFromInteraction(i)

	err := helpers.RespondImmediately(s, i, "Processing your request...")
	if err != nil {
		log.WithFields(meta).Errorf("Error responding to user: %s", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)

	pgConn, err := postgres.NewPostgresConnection()
	if err != nil {
		log.WithFields(meta).Error(err)
		err2 := helpers.RespondDelayed(s, i, "An unexpected error occurred. Please try again later :(")
		if err2 != nil {
			log.WithFields(meta).Error(err2)
			cancel()
			return
		}
		cancel()
		return
	}

	blacklistService := domain.NewBlacklistService(persistence.NewPostgresBlacklistRepository(pgConn))
	preferenceService := domain.NewPreferenceService(persistence.NewPostgresPreferenceRepository(pgConn))
	trackService := domain.NewTrackService(persistence.NewPostgresTrackRepository(pgConn))

	spClient, err := spotify.NewSpotifyClient(ctx)
	if err != nil {
		log.WithFields(meta).Error(err)
		err2 := helpers.RespondDelayed(s, i, "An unexpected error occurred. Please try again later :(")
		if err2 != nil {
			log.WithFields(meta).Error(err2)
			cancel()
			return
		}
		cancel()
		return
	}
	spTrackService := domain.NewSpotifyTrackService(&sp.SpotifyTrackRepo{
		Client: spClient,
	})
	spArtistService := domain.NewSpotifyArtistService(&sp.SpotifyArtistRepo{
		Client: spClient,
	})
	spPlaylistService := domain.NewSpotifyPlaylistService(&sp.SpotifyPlaylistRepo{
		Client: spClient,
	})

	input := &AddTrackInput{
		TrackId:           trackId,
		UserFields:        userFields,
		Ctx:               ctx,
		Meta:              meta,
		TrackService:      trackService,
		BlacklistService:  blacklistService,
		SpTrackService:    spTrackService,
		SpArtistService:   spArtistService,
		SpPlaylistService: spPlaylistService,
		PreferenceService: preferenceService,
	}
	track, err := AddTrack(input)
	cancel()

	if err != nil {
		response := "An unexpected error occurred. Please try again later :("
		switch {
		case errors.Is(err, domain.ErrInvalidTrackId):
			response = "I can't recognize that track ID!"
			break
		case errors.Is(err, domain.ErrTrackAlreadyInPlaylist):
			response = "That track is already in the playlist!"
			break
		case errors.Is(err, domain.ErrTrackTooLong):
			response = "That track is too long!"
			break
		case errors.Is(err, domain.ErrNoTrackExists):
			response = "That track does not exist!"
			break
		case errors.Is(err, domain.ErrCouldNotAddToPlaylist):
		case errors.Is(err, domain.ErrCouldNotAddToDatabase):
		case errors.Is(err, domain.ErrCouldNotRemoveFromPlaylist):
			response = "Could not add track to playlist. Please try again later :("
			break
		}

		log.WithFields(meta).Error(err)
		err2 := helpers.RespondDelayed(s, i, response)
		if err2 != nil {
			log.WithFields(meta).Error(err2)
			return
		}
		return
	}

	trackDetails := fmt.Sprintf("%s by %s", track.Name, track.Artists[0].Name)
	err = helpers.RespondDelayed(s, i, fmt.Sprintf("Added %s to the playlist!", trackDetails))
	if err != nil {
		log.WithFields(meta).Error(err)
		return
	}
}
