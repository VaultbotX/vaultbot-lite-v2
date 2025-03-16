package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
	"github.com/vaultbotx/vaultbot-lite/internal/domain"
	"time"
)

type PreferenceRepo struct {
	db *sqlx.DB
}

func NewPostgresPreferenceRepository(db *sqlx.DB) *PreferenceRepo {
	return &PreferenceRepo{
		db: db,
	}
}

type PreferenceRecord struct {
	Key   string          `db:"key"`
	Value json.RawMessage `db:"value"`
}

func (p PreferenceRepo) Set(ctx context.Context, preferenceKey domain.PreferenceKey, value any) error {
	now := time.Now().UTC()
	// update the existing preference to have an end date of now
	_, err := p.db.ExecContext(ctx, `
		WITH current_record_id AS (
			SELECT id
			FROM preferences
			WHERE key = $1
			ORDER BY id DESC
			LIMIT 1
		)
		UPDATE preferences
		SET end_time = $2
		WHERE key = $1
	`, preferenceKey, now)
	if err != nil {
		return err
	}

	_, err = p.db.ExecContext(ctx, `
		INSERT INTO preferences (key, value, start_time)
		VALUES ($1, $2, $3)
	`, preferenceKey, value, now)
	if err != nil {
		return err
	}

	return nil
}

func (p PreferenceRepo) Get(ctx context.Context, preferenceKey domain.PreferenceKey) (*domain.Preference, error) {
	var preference PreferenceRecord
	err := p.db.GetContext(ctx, &preference, `
		SELECT key, value
		FROM preferences
		WHERE key = $1
		ORDER BY id DESC
	`, preferenceKey)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &domain.Preference{
		Key:   domain.PreferenceKey(preference.Key),
		Value: preference.Value,
	}, nil
}

func (p PreferenceRepo) GetAll(ctx context.Context) (map[domain.PreferenceKey]domain.Preference, error) {
	preferences := make(map[domain.PreferenceKey]domain.Preference)
	rows, err := p.db.QueryxContext(ctx, `
		SELECT key, value
		FROM preferences
		WHERE end_time IS NULL
	`)
	if err != nil {
		return nil, err
	}
	defer func(rows *sqlx.Rows) {
		err := rows.Close()
		if err != nil {
			log.Errorf("Error closing rows: %s", err)
			return
		}
	}(rows)

	for rows.Next() {
		var preference PreferenceRecord
		err := rows.StructScan(&preference)
		if err != nil {
			return nil, err
		}

		preferences[domain.PreferenceKey(preference.Key)] = domain.Preference{
			Key:   domain.PreferenceKey(preference.Key),
			Value: preference.Value,
		}
	}

	return preferences, nil
}
