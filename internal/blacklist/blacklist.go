package blacklist

import (
	"context"
	"errors"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
	"github.com/vaultbotx/vaultbot-lite/internal/discord/commands"
	"github.com/vaultbotx/vaultbot-lite/internal/domain"
	"github.com/vaultbotx/vaultbot-lite/internal/persistence"
	mg "github.com/vaultbotx/vaultbot-lite/internal/persistence/mongo"
	"github.com/vaultbotx/vaultbot-lite/internal/types"
	"github.com/vaultbotx/vaultbot-lite/internal/utils"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

func blacklist(s *discordgo.Session, i *discordgo.InteractionCreate, isBlacklist bool) {
	meta := utils.GetFieldsFromInteraction(i)
	err := commands.CheckUserPermissions(s, i)
	if err != nil {
		if errors.Is(err, types.ErrUnauthorized) {
			err := commands.Respond(s, i, "You are not authorized to use this command")
			if err != nil {
				log.WithFields(meta).Errorf("Error responding to unauthorized user: %s", err)
				return
			}
			return
		}

		log.WithFields(meta).Errorf("Error checking user permissions: %s", err)
		err := commands.Respond(s, i, "There was an error checking your permissions")
		if err != nil {
			log.WithFields(meta).Errorf("Error responding to user: %s", err)
			return
		}
		return
	}

	selectedOption := i.ApplicationCommandData().Options[0]

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)

	instance, err := mg.GetMongoClient(ctx)
	if err != nil {
		cancel()
		err := commands.Respond(s, i, "An unexpected error occurred. Please try again later :(")
		if err != nil {
			log.WithFields(meta).Errorf("Error responding to user: %s", err)
			return
		}
		log.WithFields(meta).Errorf("Error getting MongoDB client: %s", err)
		return
	}
	defer func(instance *mongo.Client, ctx context.Context) {
		err := instance.Disconnect(ctx)
		if err != nil {
			log.Errorf("Error disconnecting from MongoDB: %v", err)
			return
		}
	}(instance, ctx)
	blacklistRepository := persistence.NewBlacklistRepository(instance)
	blacklistService := domain.NewBlacklistService(blacklistRepository)

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
		if errors.Is(err, types.ErrBlacklistItemAlreadyExists) {
			err := commands.Respond(s, i, "That item is already blacklisted!")
			if err != nil {
				log.WithFields(meta).Errorf("Error responding to user: %s", err)
				return
			}
			return
		}

		log.WithFields(meta).Errorf("Error blacklisting item: %s", err)
		err := commands.Respond(s, i, "There was an error blacklisting that item")
		if err != nil {
			log.WithFields(meta).Errorf("Error responding to user: %s", err)
			return
		}
		return
	}

	err = commands.Respond(s, i, "Done!")
	if err != nil {
		log.WithFields(meta).Errorf("Error responding to user: %s", err)
		return
	}
}
