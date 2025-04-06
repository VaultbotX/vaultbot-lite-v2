package utils

import (
	"github.com/bwmarrin/discordgo"
	"github.com/vaultbotx/vaultbot-lite/internal/domain"
)

func GetUserFieldsFromInteraction(i *discordgo.InteractionCreate) *domain.UserFields {
	isDirectMessage := "" == i.GuildID

	if isDirectMessage {
		return &domain.UserFields{
			UserId:   i.User.ID,
			Username: i.User.Username,
			GuildId:  "", // No guild ID in direct messages
		}
	}

	return &domain.UserFields{
		UserId:   i.Member.User.ID,
		Username: i.Member.User.Username,
		GuildId:  i.GuildID,
	}
}
