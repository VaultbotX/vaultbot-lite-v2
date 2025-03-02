package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/vaultbotx/vaultbot-lite/internal/domain"
)

func CheckUserPermissions(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	perms, err := s.State.UserChannelPermissions(i.Member.User.ID, i.ChannelID)
	if err != nil {
		return err
	}

	if perms&discordgo.PermissionAdministrator == 0 {
		return domain.ErrUnauthorized
	}

	return nil
}
