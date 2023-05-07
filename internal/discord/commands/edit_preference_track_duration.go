package commands

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
	"github.com/vaultbotx/vaultbot-lite/internal/commands"
	"github.com/vaultbotx/vaultbot-lite/internal/utils"
)

func EditPreferenceTrackDuration(s *discordgo.Session, i *discordgo.InteractionCreate) {
	durationInMinutes := i.ApplicationCommandData().Options[0].IntValue()
	durationInMilliseconds := int(durationInMinutes * 60 * 1000)

	meta := utils.GetFieldsFromInteraction(i)

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
