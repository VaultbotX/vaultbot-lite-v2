package blacklist

import (
	"github.com/bwmarrin/discordgo"
)

func BlacklistCommandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	blacklist(s, i, true)
}

func UnblacklistCommandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	blacklist(s, i, false)
}
