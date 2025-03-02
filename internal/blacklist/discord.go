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
			commands.Respond(s, i, "You are not authorized to use this command")
			return
		}

		log.WithFields(meta).Errorf("Error checking user permissions: %s", err)
		commands.Respond(s, i, "There was an error checking your permissions")
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
			commands.Respond(s, i, "That item is already blacklisted!")
			return
		}

		log.WithFields(meta).Errorf("Error blacklisting item: %s", err)
		commands.Respond(s, i, "There was an error blacklisting that item")
		return
	}

	commands.Respond(s, i, "Done!")
}
