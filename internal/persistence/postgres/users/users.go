package users

import (
	"github.com/vaultbotx/vaultbot-lite/internal/types"
	"time"
)

type User struct {
	Id              int       `db:"id"`
	DiscordId       string    `db:"discord_id"`
	DiscordUsername string    `db:"discord_username"`
	CreatedAt       time.Time `db:"created_at"`
}

func AddUser(fields *types.UserFields) {

}
