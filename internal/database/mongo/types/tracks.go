package types

type CommonFields struct {
	Id        string `bson:"_id"`
	GuildId   string `bson:"guildId"`
	Timestamp int64  `bson:"timestamp"`
}

type AddedTrack struct {
	TrackId  string `bson:"trackId"`
	UserId   string `bson:"userId"`
	Username string `bson:"username"`
	CommonFields
}

type BlacklistedTrack struct {
	TrackId           string `bson:"trackId"`
	BlockedById       string `bson:"blockedById"`
	BlockedByUsername string `bson:"blockedByUsername"`
	CommonFields
}

type BlacklistedArtist struct {
	ArtistId          string `bson:"artistId"`
	BlockedById       string `bson:"blockedById"`
	BlockedByUsername string `bson:"blockedByUsername"`
	CommonFields
}

type BlacklistedGenre struct {
	GenreName         string `bson:"genreName"`
	BlockedById       string `bson:"blockedById"`
	BlockedByUsername string `bson:"blockedByUsername"`
	CommonFields
}
