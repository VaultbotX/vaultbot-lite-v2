package commands

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
	"github.com/vaultbotx/vaultbot-lite/internal/commands"
	"github.com/vaultbotx/vaultbot-lite/internal/utils"
)

func EditPreferences(s *discordgo.Session, i *discordgo.InteractionCreate) {
	selectedOption := i.ApplicationCommandData().Options[0]
	meta := utils.GetFieldsFromInteraction(i)

	switch selectedOption.Name {
	case "max-track-duration":
		editPreferenceTrackDuration(s, i, selectedOption, meta)
	case "purge-frequency":
		editPreferencePurgeFrequency(s, i, selectedOption, meta)
	case "max-track-age":
		editPreferenceMaxTrackAge(s, i, selectedOption, meta)
	default:
		log.WithFields(meta).Errorf("Unknown preference option: %s", selectedOption.Name)
		respond(s, i, "Unknown preference option")
	}
}

func editPreferenceTrackDuration(s *discordgo.Session, i *discordgo.InteractionCreate,
	option *discordgo.ApplicationCommandInteractionDataOption, meta log.Fields) {
	durationInMinutes := option.IntValue()
	durationInMilliseconds := int(durationInMinutes * 60 * 1000)

	log.WithFields(meta).Infof("Setting max song duration preference to %d", durationInMilliseconds)
	err := commands.SetMaxSongDurationPreference(durationInMilliseconds)
	if err != nil {
		log.WithFields(meta).Errorf("Error setting max song duration preference: %s", err)
		respond(s, i, "There was an error setting the track duration preference")
		return
	}

	log.WithFields(meta).Infof("Max song duration preference set to %d", durationInMilliseconds)
	response := fmt.Sprintf("Max song duration preference set to %d minutes", durationInMinutes)
	respond(s, i, response)
}

func editPreferencePurgeFrequency(s *discordgo.Session, i *discordgo.InteractionCreate,
	option *discordgo.ApplicationCommandInteractionDataOption, meta log.Fields) {
	frequencyInMinutes := option.IntValue()
	frequencyInMilliseconds := int(frequencyInMinutes * 60 * 1000)

	log.WithFields(meta).Infof("Setting purge frequency preference to %d", frequencyInMilliseconds)
	err := commands.SetPurgeFrequencyPreference(frequencyInMilliseconds)
	if err != nil {
		log.WithFields(meta).Errorf("Error setting purge frequency preference: %s", err)
		respond(s, i, "There was an error setting the purge frequency preference")
		return
	}

	log.WithFields(meta).Infof("Purge frequency preference set to %d", frequencyInMilliseconds)
	response := fmt.Sprintf("Purge frequency preference set to %d minutes", frequencyInMinutes)
	respond(s, i, response)
}

func editPreferenceMaxTrackAge(s *discordgo.Session, i *discordgo.InteractionCreate,
	option *discordgo.ApplicationCommandInteractionDataOption, meta log.Fields) {
	ageInMinutes := option.IntValue()
	ageInMilliseconds := int(ageInMinutes * 60 * 1000)

	log.WithFields(meta).Infof("Setting max track age preference to %d", ageInMilliseconds)
	err := commands.SetMaxTrackAgePreference(ageInMilliseconds)
	if err != nil {
		log.WithFields(meta).Errorf("Error setting max track age preference: %s", err)
		respond(s, i, "There was an error setting the max track age preference")
		return
	}

	log.WithFields(meta).Infof("Max track age preference set to %d", ageInMilliseconds)
	response := fmt.Sprintf("Max track age preference set to %d minutes", ageInMinutes)
	respond(s, i, response)
}
