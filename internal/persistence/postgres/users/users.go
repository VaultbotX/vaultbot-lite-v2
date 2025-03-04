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
	var addUser User
	err := tx.QueryRowx(`
		INSERT INTO users (discord_id, discord_username) 
		VALUES ($1, $2)
		ON CONFLICT (discord_id) DO NOTHING
		RETURNING id, created_at
	`, fields.UserId, fields.Username).StructScan(&addUser)

	if err != nil {
		return User{}, err
	}

	addUser.DiscordId = fields.UserId
	addUser.DiscordUsername = fields.Username

	return addUser, nil
}
