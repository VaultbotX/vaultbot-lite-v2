package users

import (
	"github.com/jmoiron/sqlx"
	"github.com/vaultbotx/vaultbot-lite/internal/domain"
	"time"
)

type User struct {
	Id              int       `db:"id"`
	DiscordId       string    `db:"discord_id"`
	DiscordUsername string    `db:"discord_username"`
	CreatedAt       time.Time `db:"created_at"`
}

// AddUser adds a user to the database
func AddUser(tx *sqlx.Tx, fields *domain.UserFields) (User, error) {
	row, err := tx.NamedExec(`
		INSERT INTO users (discord_id, discord_username) 
		VALUES (:discord_id, :discord_username)
		ON CONFLICT (discord_id) DO NOTHING
	`, map[string]any{
		"discord_id":       fields.UserId,
		"discord_username": fields.Username,
	})

	if err != nil {
		return User{}, err
	}

	id, err := row.LastInsertId()
	if err != nil {
		return User{}, err
	}

	return User{
		Id:              int(id),
		DiscordId:       fields.UserId,
		DiscordUsername: fields.Username,
		CreatedAt:       time.Now(),
	}, nil
}
