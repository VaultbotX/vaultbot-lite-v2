package types

type AddedTrack struct {
	Id        string `bson:"_id"`
	TrackId   string `bson:"trackId"`
	UserId    string `bson:"userId"`
	Username  string `bson:"username"`
	GuildId   string `bson:"guildId"`
	Timestamp int64  `bson:"timestamp"`
}
