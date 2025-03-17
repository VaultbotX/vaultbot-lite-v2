package helpers

import (
	"errors"
	"github.com/bwmarrin/discordgo"
	"github.com/vaultbotx/vaultbot-lite/internal/domain"
	"os"
)

func CheckUserPermissions(s *discordgo.Session, i *discordgo.InteractionCreate) error {
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
