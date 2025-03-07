package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
	"github.com/vaultbotx/vaultbot-lite/internal/domain"
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
	_, err := p.db.ExecContext(ctx, `
		INSERT INTO preferences (key, value)
		VALUES ($1, $2)
		ON CONFLICT (key) DO UPDATE SET value = $2
	`, preferenceKey, value)
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
