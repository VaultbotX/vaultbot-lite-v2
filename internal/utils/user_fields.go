package utils

import (
	"github.com/bwmarrin/discordgo"
	"github.com/vaultbotx/vaultbot-lite/internal/domain"
)

func GetUserFieldsFromInteraction(i *discordgo.InteractionCreate) *domain.UserFields {
	return &domain.UserFields{
		UserId:   i.Member.User.ID,
		Username: i.Member.User.Username,
		GuildId:  i.GuildID,
	}
}
