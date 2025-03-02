package preferences

import (
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
	"github.com/vaultbotx/vaultbot-lite/internal/discord/commands"
	"github.com/vaultbotx/vaultbot-lite/internal/types"
	"github.com/vaultbotx/vaultbot-lite/internal/utils"
)

func EditPreferencesCommandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	selectedOption := i.ApplicationCommandData().Options[0]
	meta := utils.GetFieldsFromInteraction(i)
	err := commands.CheckUserPermissions(s, i)
	if err != nil {
		if errors.Is(err, types.ErrUnauthorized) {
			commands.Respond(s, i, "You are not authorized to use this command")
			return
		}

		log.WithFields(meta).Errorf("Error checking user permissions: %s", err)
		commands.Respond(s, i, "There was an error checking your permissions")
		return
	}

	switch selectedOption.Name {
	case "max-track-duration":
		editPreferenceTrackDuration(s, i, selectedOption, meta)
	case "purge-frequency":
		editPreferencePurgeFrequency(s, i, selectedOption, meta)
	case "max-track-age":
		editPreferenceMaxTrackAge(s, i, selectedOption, meta)
	default:
		log.WithFields(meta).Errorf("Unknown preference option: %s", selectedOption.Name)
		commands.Respond(s, i, "Exactly one option must be provided!")
	}
}

func editPreferenceTrackDuration(s *discordgo.Session, i *discordgo.InteractionCreate,
	option *discordgo.ApplicationCommandInteractionDataOption, meta log.Fields) {
	durationInMinutes := option.IntValue()
	durationInMilliseconds := int(durationInMinutes * 60 * 1000)

	log.WithFields(meta).Infof("Setting max song duration preference to %d", durationInMilliseconds)
	err := SetMaxSongDurationPreference(durationInMilliseconds)
	if err != nil {
		log.WithFields(meta).Errorf("Error setting max song duration preference: %s", err)
		commands.Respond(s, i, "There was an error setting the track duration preference")
		return
	}

	log.WithFields(meta).Infof("Max song duration preference set to %d", durationInMilliseconds)
	response := fmt.Sprintf("Max song duration preference set to %d minutes", durationInMinutes)
	commands.Respond(s, i, response)
}

func editPreferencePurgeFrequency(s *discordgo.Session, i *discordgo.InteractionCreate,
	option *discordgo.ApplicationCommandInteractionDataOption, meta log.Fields) {
	frequencyInDays := option.IntValue()
	frequencyInMilliseconds := int(frequencyInDays * 24 * 60 * 60 * 1000)

	log.WithFields(meta).Infof("Setting purge frequency preference to %d", frequencyInMilliseconds)
	err := SetPurgeFrequencyPreference(frequencyInMilliseconds)
	if err != nil {
		log.WithFields(meta).Errorf("Error setting purge frequency preference: %s", err)
		commands.Respond(s, i, "There was an error setting the purge frequency preference")
		return
	}

	log.WithFields(meta).Infof("Purge frequency preference set to %d", frequencyInMilliseconds)
	response := fmt.Sprintf("Purge frequency preference set to %d days", frequencyInDays)
	commands.Respond(s, i, response)
}

func editPreferenceMaxTrackAge(s *discordgo.Session, i *discordgo.InteractionCreate,
	option *discordgo.ApplicationCommandInteractionDataOption, meta log.Fields) {
	ageInMinutes := option.IntValue()
	ageInMilliseconds := int(ageInMinutes * 60 * 1000)

	log.WithFields(meta).Infof("Setting max track age preference to %d", ageInMilliseconds)
	err := SetMaxTrackAgePreference(ageInMilliseconds)
	if err != nil {
		log.WithFields(meta).Errorf("Error setting max track age preference: %s", err)
		commands.Respond(s, i, "There was an error setting the max track age preference")
		return
	}

	log.WithFields(meta).Infof("Max track age preference set to %d", ageInMilliseconds)
	response := fmt.Sprintf("Max track age preference set to %d minutes", ageInMinutes)
	commands.Respond(s, i, response)
}
