package preferences

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/vaultbotx/vaultbot-lite/internal/domain"
	"github.com/vaultbotx/vaultbot-lite/internal/persistence"
	"github.com/vaultbotx/vaultbot-lite/internal/persistence/postgres"
)

func CheckDefaultPreferences(ctx context.Context) error {
	pgConn, err := postgres.NewPostgresConnection()
	if err != nil {
		return err
	}

	preferenceService := domain.NewPreferenceService(persistence.NewPostgresPreferenceRepository(pgConn))

	preferences, err := preferenceService.Repo.GetAll(ctx)
	if err != nil {
		return err
	}

	for _, preferenceKey := range domain.AllPreferences {
		if _, ok := preferences[preferenceKey]; !ok {
			log.Infof("Preference %s does not exist, creating with default value", preferenceKey)
			err = preferenceService.Repo.Set(ctx, preferenceKey, preferenceKey.DefaultValue())
			if err != nil {
				return err
			}
		}
	}

	return nil
}
