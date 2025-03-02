package preferences

import (
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
	"github.com/vaultbotx/vaultbot-lite/internal/discord/helpers"
	"github.com/vaultbotx/vaultbot-lite/internal/domain"
	"github.com/vaultbotx/vaultbot-lite/internal/utils"
)

func EditPreferencesCommandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	selectedOption := i.ApplicationCommandData().Options[0]
	meta := utils.GetFieldsFromInteraction(i)
	err := helpers.CheckUserPermissions(s, i)
	if err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			err := helpers.Respond(s, i, "You are not authorized to use this command")
			if err != nil {
				log.WithFields(meta).Errorf("Error responding to unauthorized user: %s", err)
				return
			}
			return
		}

		log.WithFields(meta).Errorf("Error checking user permissions: %s", err)
		err := helpers.Respond(s, i, "There was an error checking your permissions")
		if err != nil {
			log.WithFields(meta).Errorf("Error responding to user: %s", err)
			return
		}
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
		err := helpers.Respond(s, i, "Exactly one option must be provided!")
		if err != nil {
			log.WithFields(meta).Errorf("Error responding to user: %s", err)
			return
		}
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
		err := helpers.Respond(s, i, "There was an error setting the track duration preference")
		if err != nil {
			log.WithFields(meta).Errorf("Error responding to user: %s", err)
			return
		}
		return
	}

	log.WithFields(meta).Infof("Max song duration preference set to %d", durationInMilliseconds)
	response := fmt.Sprintf("Max song duration preference set to %d minutes", durationInMinutes)
	err = helpers.Respond(s, i, response)
	if err != nil {
		log.WithFields(meta).Errorf("Error responding to user: %s", err)
		return
	}
}

func editPreferencePurgeFrequency(s *discordgo.Session, i *discordgo.InteractionCreate,
	option *discordgo.ApplicationCommandInteractionDataOption, meta log.Fields) {
	frequencyInDays := option.IntValue()
	frequencyInMilliseconds := int(frequencyInDays * 24 * 60 * 60 * 1000)

	log.WithFields(meta).Infof("Setting purge frequency preference to %d", frequencyInMilliseconds)
	err := SetPurgeFrequencyPreference(frequencyInMilliseconds)
	if err != nil {
		log.WithFields(meta).Errorf("Error setting purge frequency preference: %s", err)
		err := helpers.Respond(s, i, "There was an error setting the purge frequency preference")
		if err != nil {
			log.WithFields(meta).Errorf("Error responding to user: %s", err)
			return
		}
		return
	}

	log.WithFields(meta).Infof("Purge frequency preference set to %d", frequencyInMilliseconds)
	response := fmt.Sprintf("Purge frequency preference set to %d days", frequencyInDays)
	err = helpers.Respond(s, i, response)
	if err != nil {
		log.WithFields(meta).Errorf("Error responding to user: %s", err)
		return
	}
}

func editPreferenceMaxTrackAge(s *discordgo.Session, i *discordgo.InteractionCreate,
	option *discordgo.ApplicationCommandInteractionDataOption, meta log.Fields) {
	ageInMinutes := option.IntValue()
	ageInMilliseconds := int(ageInMinutes * 60 * 1000)

	log.WithFields(meta).Infof("Setting max track age preference to %d", ageInMilliseconds)
	err := SetMaxTrackAgePreference(ageInMilliseconds)
	if err != nil {
		log.WithFields(meta).Errorf("Error setting max track age preference: %s", err)
		err := helpers.Respond(s, i, "There was an error setting the max track age preference")
		if err != nil {
			log.WithFields(meta).Errorf("Error responding to user: %s", err)
			return
		}
		return
	}

	log.WithFields(meta).Infof("Max track age preference set to %d", ageInMilliseconds)
	response := fmt.Sprintf("Max track age preference set to %d minutes", ageInMinutes)
	err = helpers.Respond(s, i, response)
	if err != nil {
		log.WithFields(meta).Errorf("Error responding to user: %s", err)
		return
	}
}
