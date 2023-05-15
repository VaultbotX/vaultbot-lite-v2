package types

type AddedTrack struct {
	Id        string `bson:"_id"`
	TrackId   string `bson:"trackId"`
	UserId    string `bson:"userId"`
	Username  string `bson:"username"`
	GuildId   string `bson:"guildId"`
	Timestamp int64  `bson:"timestamp"`
}

type BlacklistedTrack struct {
	Id                string `bson:"_id"`
	TrackId           string `bson:"trackId"`
	BlockedById       string `bson:"blockedById"`
	BlockedByUsername string `bson:"blockedByUsername"`
	GuildId           string `bson:"guildId"`
	Timestamp         int64  `bson:"timestamp"`
}

type BlacklistedArtist struct {
	Id                string `bson:"_id"`
	ArtistId          string `bson:"artistId"`
	BlockedById       string `bson:"blockedById"`
	BlockedByUsername string `bson:"blockedByUsername"`
	GuildId           string `bson:"guildId"`
	Timestamp         int64  `bson:"timestamp"`
}

type BlacklistedGenre struct {
	Id                string `bson:"_id"`
	GenreName         string `bson:"genreName"`
	BlockedById       string `bson:"blockedById"`
	BlockedByUsername string `bson:"blockedByUsername"`
	GuildId           string `bson:"guildId"`
	Timestamp         int64  `bson:"timestamp"`
}
