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
	"github.com/vaultbotx/vaultbot-lite/internal/utils"
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

	selectedOption := i.ApplicationCommandData().Options[0]

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

	curried := func(blacklistType domain.BlacklistType) error {
		if isBlacklist {
			return blacklistService.Repo.AddToBlacklist(ctx, blacklistType, selectedOption.StringValue(), utils.GetUserFieldsFromInteraction(i), time.Now())
		}

		return blacklistService.Repo.RemoveFromBlacklist(ctx, blacklistType, selectedOption.StringValue())
	}

	switch selectedOption.Name {
	case "track":
		err = curried(domain.Track)
	case "artist":
		err = curried(domain.Artist)
	case "genre":
		err = curried(domain.Genre)
	}
	cancel()

	if err != nil {
		if errors.Is(err, domain.ErrBlacklistItemAlreadyExists) {
			err := helpers.RespondDelayed(s, i, "That item is already blacklisted!")
			if err != nil {
				log.WithFields(meta).Errorf("Error responding to user: %s", err)
				return
			}
			return
		}

		log.WithFields(meta).Errorf("Error blacklisting item: %s", err)
		err := helpers.RespondDelayed(s, i, "There was an error blacklisting that item")
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
