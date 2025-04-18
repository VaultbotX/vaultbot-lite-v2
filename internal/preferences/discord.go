package preferences

import (
	"context"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
	"github.com/vaultbotx/vaultbot-lite/internal/discord/helpers"
	"github.com/vaultbotx/vaultbot-lite/internal/domain"
	"github.com/vaultbotx/vaultbot-lite/internal/persistence"
	"github.com/vaultbotx/vaultbot-lite/internal/persistence/postgres"
	"github.com/vaultbotx/vaultbot-lite/internal/utils"
	"time"
)

var Command = &discordgo.ApplicationCommand{
	Name:        "edit-preferences",
	Description: "Edit preferences",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Name:        "max-track-duration",
			Description: "The track duration in minutes",
			Type:        discordgo.ApplicationCommandOptionInteger,
			Required:    false,
			MinValue:    &helpers.MinSongDuration,
			MaxValue:    helpers.MaxSongDuration,
		},
		{
			Name:        "purge-frequency",
			Description: "How often to purge the playlist (in days)",
			Type:        discordgo.ApplicationCommandOptionInteger,
			Required:    false,
			MinValue:    &helpers.MinPurgeFrequency,
			MaxValue:    helpers.MaxPurgeFrequency,
		},
		{
			Name:        "max-track-age",
			Description: "The maximum age of a track in days",
			Type:        discordgo.ApplicationCommandOptionInteger,
			Required:    false,
			MinValue:    &helpers.MinTrackAge,
			MaxValue:    helpers.MaxTrackAge,
		},
	},
	DefaultMemberPermissions: &helpers.AdminPermission,
}

func EditPreferencesCommandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	selectedOption := i.ApplicationCommandData().Options[0]
	meta := utils.GetFieldsFromInteraction(i)

	log.WithFields(meta).Infof("Received edit-preferences command with option: %s", selectedOption.Name)

	err := helpers.CheckUserPermissions(s, i)
	if err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			err := helpers.RespondImmediately(s, i, "You are not authorized to use this command")
			if err != nil {
				log.WithFields(meta).Errorf("Error responding to unauthorized user: %s", err)
				return
			}
			return
		}

		log.WithFields(meta).Errorf("Error checking user permissions: %s", err)
		err := helpers.RespondImmediately(s, i, "There was an error checking your permissions")
		if err != nil {
			log.WithFields(meta).Errorf("Error responding to user: %s", err)
			return
		}
		return
	}

	err = helpers.RespondImmediately(s, i, "Processing your request...")
	if err != nil {
		log.WithFields(meta).Errorf("Error responding to user: %s", err)
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
		err := helpers.RespondDelayed(s, i, "Exactly one option must be provided!")
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

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	pgConn, err := postgres.NewPostgresConnection()
	if err != nil {
		log.WithFields(meta).Error(err)
		err2 := helpers.RespondDelayed(s, i, "An unexpected error occurred. Please try again later :(")
		if err2 != nil {
			log.WithFields(meta).Error(err2)
			cancel()
			return
		}
		cancel()
		return
	}
	preferenceService := domain.NewPreferenceService(persistence.NewPostgresPreferenceRepository(pgConn))

	log.WithFields(meta).Infof("Setting max song duration preference to %d", durationInMilliseconds)
	err = preferenceService.Repo.Set(ctx, domain.MaxDurationKey, durationInMilliseconds)
	if err != nil {
		log.WithFields(meta).Errorf("Error setting max song duration preference: %s", err)
		err2 := helpers.RespondDelayed(s, i, "There was an error setting the track duration preference")
		if err2 != nil {
			log.WithFields(meta).Errorf("Error responding to user: %s", err2)
			cancel()
			return
		}
		cancel()
		return
	}

	log.WithFields(meta).Infof("Max song duration preference set to %d", durationInMilliseconds)
	response := fmt.Sprintf("Max song duration preference set to %d minutes", durationInMinutes)
	err = helpers.RespondDelayed(s, i, response)
	if err != nil {
		log.WithFields(meta).Errorf("Error responding to user: %s", err)
		cancel()
		return
	}

	cancel()
}

func editPreferencePurgeFrequency(s *discordgo.Session, i *discordgo.InteractionCreate,
	option *discordgo.ApplicationCommandInteractionDataOption, meta log.Fields) {
	frequencyInDays := option.IntValue()
	frequencyInMilliseconds := int(frequencyInDays * 24 * 60 * 60 * 1000)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	pgConn, err := postgres.NewPostgresConnection()
	if err != nil {
		log.WithFields(meta).Error(err)
		err2 := helpers.RespondDelayed(s, i, "An unexpected error occurred. Please try again later :(")
		if err2 != nil {
			log.WithFields(meta).Error(err2)
			cancel()
			return
		}
		cancel()
		return
	}
	preferenceService := domain.NewPreferenceService(persistence.NewPostgresPreferenceRepository(pgConn))

	log.WithFields(meta).Infof("Setting purge frequency preference to %d", frequencyInMilliseconds)
	err = preferenceService.Repo.Set(ctx, domain.PurgeFrequencyKey, frequencyInMilliseconds)
	if err != nil {
		log.WithFields(meta).Errorf("Error setting purge frequency preference: %s", err)
		err2 := helpers.RespondDelayed(s, i, "There was an error setting the purge frequency preference")
		if err2 != nil {
			log.WithFields(meta).Errorf("Error responding to user: %s", err2)
			cancel()
			return
		}
		cancel()
		return
	}

	log.WithFields(meta).Infof("Purge frequency preference set to %d", frequencyInMilliseconds)
	response := fmt.Sprintf("Purge frequency preference set to %d days", frequencyInDays)
	err = helpers.RespondDelayed(s, i, response)
	if err != nil {
		log.WithFields(meta).Errorf("Error responding to user: %s", err)
		cancel()
		return
	}

	cancel()
}

func editPreferenceMaxTrackAge(s *discordgo.Session, i *discordgo.InteractionCreate,
	option *discordgo.ApplicationCommandInteractionDataOption, meta log.Fields) {
	ageInMinutes := option.IntValue()
	ageInMilliseconds := int(ageInMinutes * 60 * 1000)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	pgConn, err := postgres.NewPostgresConnection()
	if err != nil {
		log.WithFields(meta).Error(err)
		err2 := helpers.RespondDelayed(s, i, "An unexpected error occurred. Please try again later :(")
		if err2 != nil {
			log.WithFields(meta).Error(err2)
			cancel()
			return
		}
		cancel()
		return
	}
	preferenceService := domain.NewPreferenceService(persistence.NewPostgresPreferenceRepository(pgConn))

	log.WithFields(meta).Infof("Setting max track age preference to %d", ageInMilliseconds)
	err = preferenceService.Repo.Set(ctx, domain.MaxTrackAgeKey, ageInMilliseconds)
	if err != nil {
		log.WithFields(meta).Errorf("Error setting max track age preference: %s", err)
		err2 := helpers.RespondDelayed(s, i, "There was an error setting the max track age preference")
		if err2 != nil {
			log.WithFields(meta).Errorf("Error responding to user: %s", err2)
			cancel()
			return
		}
		cancel()
		return
	}

	log.WithFields(meta).Infof("Max track age preference set to %d", ageInMilliseconds)
	response := fmt.Sprintf("Max track age preference set to %d minutes", ageInMinutes)
	err = helpers.RespondDelayed(s, i, response)
	if err != nil {
		log.WithFields(meta).Errorf("Error responding to user: %s", err)
		cancel()
		return
	}

	cancel()
}
