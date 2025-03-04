package users

import (
	"database/sql"
	"errors"
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
		SELECT id, discord_id, discord_username, created_at
		FROM users
		WHERE discord_id = $1
	`, fields.UserId).StructScan(&addUser)

	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return User{}, err
		}

		// no user found, create a new one
		err := tx.QueryRowx(`
			INSERT INTO users (discord_id, discord_username) 
			VALUES ($1, $2)
			RETURNING id, created_at
		`, fields.UserId, fields.Username).StructScan(&addUser)

		if err != nil {
			return User{}, err
		}

		addUser.DiscordId = fields.UserId
		addUser.DiscordUsername = fields.Username
	}

	return addUser, nil
}
