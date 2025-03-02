package blacklist

import (
	"context"
	"errors"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
	"github.com/vaultbotx/vaultbot-lite/internal/discord/commands"
	"github.com/vaultbotx/vaultbot-lite/internal/types"
	"github.com/vaultbotx/vaultbot-lite/internal/utils"
	"time"
)

func BlacklistCommandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	blacklist(s, i, true)
}

func UnblacklistCommandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	blacklist(s, i, false)
}

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

	curried := func(blacklistType BlacklistType) error {
		if isBlacklist {
			return Blacklist(ctx, blacklistType, selectedOption.StringValue(), utils.GetUserFieldsFromInteraction(i))
		}

		return Unblacklist(ctx, blacklistType, selectedOption.StringValue(), utils.GetUserFieldsFromInteraction(i))
	}

	switch selectedOption.Name {
	case "track":
		err = curried(Track)
	case "artist":
		err = curried(Artist)
	case "genre":
		err = curried(Genre)
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
