package blacklist

import (
	"context"
	"errors"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
	"github.com/vaultbotx/vaultbot-lite/internal/discord/helpers"
	"github.com/vaultbotx/vaultbot-lite/internal/domain"
	"github.com/vaultbotx/vaultbot-lite/internal/persistence"
	"github.com/vaultbotx/vaultbot-lite/internal/persistence/postgres"
	"github.com/vaultbotx/vaultbot-lite/internal/spotify"
	"github.com/vaultbotx/vaultbot-lite/internal/utils"
	"strings"
	"time"
)

func blacklist(s *discordgo.Session, i *discordgo.InteractionCreate, isBlacklist bool) {
	meta := utils.GetFieldsFromInteraction(i)
	err := helpers.CheckUserPermissions(s, i)
	if err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			err := helpers.RespondImmediately(s, i, "You are not authorized to use this command")
			if err != nil {
				log.WithFields(meta).Errorf("Error responding to unauthorized user: %s", err)
				return
			}
			return
		}

		log.WithFields(meta).Errorf("Error checking user permissions: %s", err)
		err := helpers.RespondImmediately(s, i, "There was an error checking your permissions")
		if err != nil {
			log.WithFields(meta).Errorf("Error responding to user: %s", err)
			return
		}
		return
	}

	err = helpers.RespondImmediately(s, i, "Processing your request...")
	if err != nil {
		log.WithFields(meta).Errorf("Error responding to user: %s", err)
		return
	}

	commandData := i.ApplicationCommandData()
	selectedOption := commandData.Options[0]

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

	curried := func(blacklistType domain.EntityType, value string) error {
		if isBlacklist {
			return blacklistService.Repo.AddToBlacklist(ctx, blacklistType, value, utils.GetUserFieldsFromInteraction(i))
		}

		return blacklistService.Repo.RemoveFromBlacklist(ctx, blacklistType, value)
	}

	switch selectedOption.Name {
	case "track":
		trackId := spotify.ParseSpotifyId(selectedOption.StringValue(), domain.Track)
		if trackId == nil {
			err = domain.ErrInvalidSpotifyId
			break
		}
		err = curried(domain.Track, trackId.String())
	case "artist":
		artistId := spotify.ParseSpotifyId(selectedOption.StringValue(), domain.Artist)
		if artistId == nil {
			err = domain.ErrInvalidSpotifyId
			break
		}
		err = curried(domain.Artist, artistId.String())
	case "genre":
		genreName := strings.ToLower(selectedOption.StringValue())
		err = curried(domain.Genre, genreName)
	}
	cancel()

	if err != nil {
		response := "There was an error managing that blacklist item"
		switch {
		case errors.Is(err, domain.ErrBlacklistItemAlreadyExists):
			response = "That item is already blacklisted!"
			break
		case errors.Is(err, domain.ErrInvalidSpotifyId):
			response = "Please provide a valid ID, URI, or URL"
			break
		}

		log.WithFields(meta).Errorf("Error managing blacklist item: %s", err)
		err := helpers.RespondDelayed(s, i, response)
		if err != nil {
			log.WithFields(meta).Errorf("Error responding to user: %s", err)
			return
		}
		return
	}

	err = helpers.RespondDelayed(s, i, "Done!")
	if err != nil {
		log.WithFields(meta).Errorf("Error responding to user: %s", err)
		return
	}
}
