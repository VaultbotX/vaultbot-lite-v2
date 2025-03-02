package preferences

import (
	"context"
	log "github.com/sirupsen/logrus"
)

func CheckDefaultPreferences(ctx context.Context) error {
	preferences, err := GetAllPreferences(ctx)
	if err != nil {
		return err
	}

	for _, preferenceKey := range AllPreferences {
		if _, ok := preferences[preferenceKey]; !ok {
			log.Info("Preference %s does not exist, creating with default value", preferenceKey)
			err = SetPreference(ctx, preferenceKey, preferenceKey.DefaultValue())
			if err != nil {
				return err
			}
		}
	}

	return nil
}
