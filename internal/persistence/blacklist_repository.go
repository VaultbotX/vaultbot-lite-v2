package persistence

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
	"github.com/vaultbotx/vaultbot-lite/internal/domain"
	"time"
)

type BlacklistRepository struct {
	db *sqlx.DB
}

type BlacklistRecord struct {
	Id              int       `db:"id"`
	Type            string    `db:"type"`
	Value           string    `db:"value"`
	BlockedByUserId int       `db:"blocked_by_user_id"`
	CreatedAt       time.Time `db:"created_at"`
}

func NewPostgresBlacklistRepository(db *sqlx.DB) *BlacklistRepository {
	return &BlacklistRepository{
		db: db,
	}
}

func (r *BlacklistRepository) AddToBlacklist(ctx context.Context, blacklistType domain.EntityType, id string,
	userFields *domain.UserFields) error {
	_, err := r.db.ExecContext(ctx, `
		WITH user_id AS (
		    SELECT id
		    FROM users
		    WHERE discord_id = $1
		)
		INSERT INTO blacklist (type, value, blocked_by_user_id)
		VALUES ($2, $3, (SELECT id FROM user_id))
		ON CONFLICT (value) DO NOTHING
	`, userFields.UserId, blacklistType, id)
	if err != nil {
		return err
	}

	return nil
}

func (r *BlacklistRepository) RemoveFromBlacklist(ctx context.Context, blacklistType domain.EntityType, id string) error {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM blacklist
		WHERE type = $1 AND value = $2
	`, blacklistType, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("blacklist item not found")
	}

	return nil
}

func (r *BlacklistRepository) CheckBlacklistItem(ctx context.Context, blacklistType domain.EntityType, id string) (bool, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `
		SELECT COUNT(*)
		FROM blacklist
		WHERE type = $1 AND value = $2
	`, blacklistType, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		return false, err
	}

	return count > 0, nil
}
