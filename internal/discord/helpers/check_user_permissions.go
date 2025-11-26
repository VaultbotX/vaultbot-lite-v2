package helpers

import (
	"errors"
	"os"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
	"github.com/vaultbotx/vaultbot-lite/internal/domain"
)

func EnsureAdministratorRoleForUser(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	perms, err := s.State.UserChannelPermissions(i.Member.User.ID, i.ChannelID)
	if err != nil {
		return err
	}

	if perms&discordgo.PermissionAdministrator == 0 {
		return domain.ErrUnauthorized
	}

	// Secondary check to limit commands to only the bot owner for now
	administratorUserId, exists := os.LookupEnv("DISCORD_ADMINISTRATOR_USER_ID")
	if !exists {
		return errors.New("DISCORD_ADMINISTRATOR_USER_ID not set")
	}

	if i.Member.User.ID != administratorUserId {
		return domain.ErrUnauthorized
	}

	return nil
}

func HandleUserPermissionError(s *discordgo.Session, i *discordgo.InteractionCreate, originalError error, meta log.Fields) {
	if errors.Is(originalError, domain.ErrUnauthorized) {
		err := RespondImmediately(s, i, "You are not authorized to use this command")
		if err != nil {
			log.WithFields(meta).Errorf("Error responding to unauthorized user: %s", err)
			return
		}
		return
	}

	log.WithFields(meta).Errorf("Error checking user permissions: %s", originalError)
	err := RespondImmediately(s, i, "There was an error checking your permissions")
	if err != nil {
		log.WithFields(meta).Errorf("Error responding to user: %s", err)
		return
	}
	return
}
