package utils

import (
	"github.com/bwmarrin/discordgo"
	"github.com/vaultbotx/vaultbot-lite/internal/types"
)

func GetUserFieldsFromInteraction(i *discordgo.InteractionCreate) *types.UserFields {
	return &types.UserFields{
		UserId:   i.Member.User.ID,
		Username: i.Member.User.Username,
		GuildId:  i.GuildID,
	}
}
