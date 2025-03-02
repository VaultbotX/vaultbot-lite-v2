package commands

import "github.com/bwmarrin/discordgo"

func Respond(s *discordgo.Session, i *discordgo.InteractionCreate, response string) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: response,
		},
	})
}
