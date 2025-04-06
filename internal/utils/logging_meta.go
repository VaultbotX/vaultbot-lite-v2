package utils

import (
	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

func GetFieldsFromInteraction(i *discordgo.InteractionCreate) logrus.Fields {
	isDirectMessage := "" == i.GuildID

	if isDirectMessage {
		return logrus.Fields{
			"userId":          i.User.ID,
			"username":        i.User.Username + i.User.Discriminator,
			"isDirectMessage": isDirectMessage,
		}
	}

	return logrus.Fields{
		"userId":          i.Member.User.ID,
		"username":        i.Member.User.Username + i.Member.User.Discriminator,
		"guildId":         i.GuildID,
		"isDirectMessage": isDirectMessage,
	}
}
