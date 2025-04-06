package domain

type UserFields struct {
	UserId   string
	Username string
	// GuildId may be empty if the user is in a direct message
	GuildId string
}
