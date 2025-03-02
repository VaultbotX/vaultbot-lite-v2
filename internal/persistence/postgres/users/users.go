package users

import (
	"github.com/jmoiron/sqlx"
	"github.com/vaultbotx/vaultbot-lite/internal/types"
	"time"
)

type User struct {
	Id              int       `db:"id"`
	DiscordId       string    `db:"discord_id"`
	DiscordUsername string    `db:"discord_username"`
	CreatedAt       time.Time `db:"created_at"`
}

var DefaultUser = User{}

// AddUser adds a user to the database
func AddUser(tx *sqlx.Tx, fields *types.UserFields) (User, error) {
	row, err := tx.NamedExec(`
		INSERT INTO users (discord_id, discord_username) 
		VALUES (:discord_id, :discord_username)
		ON CONFLICT (discord_id) DO NOTHING
	`, map[string]interface{}{
		"discord_id":       fields.UserId,
		"discord_username": fields.Username,
	})

	if err != nil {
		return DefaultUser, err
	}

	id, err := row.LastInsertId()
	if err != nil {
		return DefaultUser, err
	}

	return User{
		Id:              int(id),
		DiscordId:       fields.UserId,
		DiscordUsername: fields.Username,
		CreatedAt:       time.Now(),
	}, nil
}
