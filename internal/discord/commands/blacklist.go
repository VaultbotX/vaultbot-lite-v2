package commands

import (
	"context"
	"errors"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
	"github.com/vaultbotx/vaultbot-lite/internal/commands"
	"github.com/vaultbotx/vaultbot-lite/internal/types"
	"github.com/vaultbotx/vaultbot-lite/internal/utils"
	"time"
)

func Blacklist(s *discordgo.Session, i *discordgo.InteractionCreate) {
	blacklist(s, i, true)
}

func Unblacklist(s *discordgo.Session, i *discordgo.InteractionCreate) {
	blacklist(s, i, false)
}

func blacklist(s *discordgo.Session, i *discordgo.InteractionCreate, isBlacklist bool) {
	meta := utils.GetFieldsFromInteraction(i)
	err := CheckUserPermissions(s, i)
	if err != nil {
		if errors.Is(err, types.ErrUnauthorized) {
			respond(s, i, "You are not authorized to use this command")
			return
		}

		log.WithFields(meta).Errorf("Error checking user permissions: %s", err)
		respond(s, i, "There was an error checking your permissions")
		return
	}

	selectedOption := i.ApplicationCommandData().Options[0]

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)

	curried := func(blacklistType types.BlacklistType) error {
		if isBlacklist {
			return commands.Blacklist(ctx, blacklistType, selectedOption.StringValue(), utils.GetUserFieldsFromInteraction(i))
		}

		return commands.Unblacklist(ctx, blacklistType, selectedOption.StringValue(), utils.GetUserFieldsFromInteraction(i))
	}

	switch selectedOption.Name {
	case "track":
		err = curried(types.Track)
	case "artist":
		err = curried(types.Artist)
	case "genre":
		err = curried(types.Genre)
	}
	cancel()

	if err != nil {
		if errors.Is(err, types.ErrBlacklistItemAlreadyExists) {
			respond(s, i, "That item is already blacklisted!")
			return
		}

		log.WithFields(meta).Errorf("Error blacklisting item: %s", err)
		respond(s, i, "There was an error blacklisting that item")
		return
	}

	respond(s, i, "Done!")
}
