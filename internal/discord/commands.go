package discord

import (
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
	internalcommands "github.com/vaultbotx/vaultbot-lite/internal/commands"
	"github.com/vaultbotx/vaultbot-lite/internal/types"
	"time"
)

// TODO: Hook this up and be sure to include the slash command option
func addTrack(s *discordgo.Session, i *discordgo.InteractionCreate) {
	trackId := i.ApplicationCommandData().Options[0].StringValue()

	meta := types.GetFieldsFromInteraction(i)
	track, err := internalcommands.AddTrack(context.Background(), trackId, meta)
	// TODO: Handle all of the different error types here
	if err != nil {
		log.WithFields(meta).Error(err)
		return
	}

	trackDetails := fmt.Sprintf("%s by %s", track.Name, track.Artists[0].Name)
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Added %s to the playlist!", trackDetails),
		},
	})

	if err != nil {
		log.WithFields(meta).Error(err)
		return
	}
}

func basicCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Hey there! Congratulations, you just executed your first slash command",
		},
	})

	if err != nil {
		log.Error(err)
		return
	}
}

func subcommands(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	content := ""

	// As you can see, names of subcommands (nested, top-level)
	// and subcommand groups are provided through the arguments.
	switch options[0].Name {
	case "subcommand":
		content = "The top-level subcommand is executed. Now try to execute the nested one."
	case "subcommand-group":
		options = options[0].Options
		switch options[0].Name {
		case "nested-subcommand":
			content = "Nice, now you know how to execute nested commands too"
		default:
			content = "Oops, something went wrong.\n" +
				"Hol' up, you aren't supposed to see this message."
		}
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})

	if err != nil {
		log.Error(err)
		return
	}
}

func responses(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Responses to a command are very important.
	// First of all, because you need to react to the interaction
	// by sending the response in 3 seconds after receiving, otherwise
	// interaction will be considered invalid, and you can no longer
	// use the interaction token and ID for responding to the user's request

	content := ""
	// As you can see, the response type names used here are pretty self-explanatory,
	// but for those who want more information see the official documentation
	switch i.ApplicationCommandData().Options[0].IntValue() {
	case int64(discordgo.InteractionResponseChannelMessageWithSource):
		content =
			"You just responded to an interaction, sent a message and showed the original one. " +
				"Congratulations!"
		content +=
			"\nAlso... you can edit your response, wait 5 seconds and this message will be changed"
	default:
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseType(i.ApplicationCommandData().Options[0].IntValue()),
		})

		if err != nil {
			log.Error(err)
			_, err2 := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: "Something went wrong",
			})

			if err2 != nil {
				log.Error(err2)
				return
			}
		}

		return
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseType(i.ApplicationCommandData().Options[0].IntValue()),
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})
	if err != nil {
		log.Error(err)
		_, err2 := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "Something went wrong",
		})

		if err2 != nil {
			log.Error(err2)
			return
		}

		return
	}
	time.AfterFunc(time.Second*5, func() {
		content := content + "\n\nWell, now you know how to create and edit responses. " +
			"But you still don't know how to delete them... so... wait 10 seconds and this " +
			"message will be deleted."
		_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		if err != nil {
			log.Error(err)

			_, err2 := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: "Something went wrong",
			})

			if err != nil {
				log.Error(err2)
				return
			}

			return
		}

		time.Sleep(time.Second * 10)
		err := s.InteractionResponseDelete(i.Interaction)
		if err != nil {
			log.Error(err)
			return
		}
	})
}

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name: "basic-command",
			// All commands and options must have a description
			// Commands/options without description will fail the registration
			// of the command.
			Description: "Basic command",
		},
		{
			Name:        "subcommands",
			Description: "Subcommands and command groups example",
			Options: []*discordgo.ApplicationCommandOption{
				// When a command has subcommands/subcommand groups
				// It must not have top-level options, they aren't accessible in the UI
				// in this case (at least not yet), so if a command has
				// subcommands/subcommand any groups registering top-level options
				// will cause the registration of the command to fail

				{
					Name:        "subcommand-group",
					Description: "Subcommands group",
					Options: []*discordgo.ApplicationCommandOption{
						// Also, subcommand groups aren't capable of
						// containing options, by the name of them, you can see
						// they can only contain subcommands
						{
							Name:        "nested-subcommand",
							Description: "Nested subcommand",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
						},
					},
					Type: discordgo.ApplicationCommandOptionSubCommandGroup,
				},
				// Also, you can create both subcommand groups and subcommands
				// in the command at the same time. But, there's some limits to
				// nesting, count of subcommands (top level and nested) and options.
				// Read the intro of slash-commands docs on Discord dev portal
				// to get more information
				{
					Name:        "subcommand",
					Description: "Top-level subcommand",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
			},
		},
		{
			Name:        "responses",
			Description: "Interaction responses testing initiative",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "resp-type",
					Description: "Response type",
					Type:        discordgo.ApplicationCommandOptionInteger,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "Channel message with source",
							Value: 4,
						},
						{
							Name:  "Deferred response With Source",
							Value: 5,
						},
					},
					Required: true,
				},
			},
		},
	}
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"basic-command": basicCommand,
		"subcommands":   subcommands,
		"responses":     responses,
	}
)
