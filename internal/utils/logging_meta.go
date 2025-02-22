package utils

import (
	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

func GetFieldsFromInteraction(interaction *discordgo.InteractionCreate) logrus.Fields {
	return logrus.Fields{
		"userId":   interaction.Member.User.ID,
		"username": interaction.Member.User.Username + interaction.Member.User.Discriminator,
		"guildId":  interaction.GuildID,
	}
}
