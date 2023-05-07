package types

type AddedTrack struct {
	TrackId   string `bson:"trackId"`
	UserId    string `bson:"userId"`
	Username  string `bson:"username"`
	GuildId   string `bson:"guildId"`
	Timestamp int64  `bson:"timestamp"`
}
