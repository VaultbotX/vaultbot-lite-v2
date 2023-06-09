package commands

import (
	"context"
	mg "github.com/vaultbotx/vaultbot-lite/internal/database/mongo"
	"github.com/vaultbotx/vaultbot-lite/internal/database/mongo/types"
	internaltypes "github.com/vaultbotx/vaultbot-lite/internal/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type BlacklistType int

const (
	Track BlacklistType = iota
	Artist
	Genre
)

func Blacklist(ctx context.Context, blacklistType BlacklistType, id string,
	userFields *internaltypes.UserFields, time time.Time) error {
	instance, err := mg.GetMongoClient(ctx)
	if err != nil {
		return err
	}
	defer instance.Disconnect(ctx)

	collection := instance.Database(mg.DatabaseName).Collection(mg.BlacklistCollection)
	var blacklistedItem interface{}
	switch blacklistType {
	case Track:
		blacklistedItem = types.BlacklistedTrack{
			TrackId:           id,
			BlockedById:       userFields.UserId,
			BlockedByUsername: userFields.Username,
			GuildId:           userFields.GuildId,
			Timestamp:         time.Unix(),
		}
	case Artist:
		blacklistedItem = types.BlacklistedArtist{
			ArtistId:          id,
			BlockedById:       userFields.UserId,
			BlockedByUsername: userFields.Username,
			GuildId:           userFields.GuildId,
			Timestamp:         time.Unix(),
		}
	case Genre:
		blacklistedItem = types.BlacklistedGenre{
			GenreName:         id,
			BlockedById:       userFields.UserId,
			BlockedByUsername: userFields.Username,
			GuildId:           userFields.GuildId,
			Timestamp:         time.Unix(),
		}
	}

	_, err = collection.InsertOne(ctx, blacklistedItem)
	if err != nil {
		return err
	}

	return nil
}

func Unblacklist(ctx context.Context, blacklistType BlacklistType, id string,
	userFields *internaltypes.UserFields) error {
	instance, err := mg.GetMongoClient(ctx)
	if err != nil {
		return err
	}
	defer instance.Disconnect(ctx)

	collection := instance.Database(mg.DatabaseName).Collection(mg.BlacklistCollection)

	var filter bson.M
	switch blacklistType {
	case Track:
		filter = bson.M{"trackId": id}
	case Artist:
		filter = bson.M{"artistId": id}
	case Genre:
		filter = bson.M{"genreName": id}
	}

	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return types.ErrNoDocuments
	}

	return nil
}

func CheckBlacklistItem(ctx context.Context, blacklistType BlacklistType, id string) (bool, error) {
	instance, err := mg.GetMongoClient(ctx)
	if err != nil {
		return false, err
	}
	defer instance.Disconnect(ctx)

	collection := instance.Database(mg.DatabaseName).Collection(mg.BlacklistCollection)

	var filter bson.M
	switch blacklistType {
	case Track:
		filter = bson.M{"trackId": id}
	case Artist:
		filter = bson.M{"artistId": id}
	case Genre:
		filter = bson.M{"genreName": id}
	}

	result := collection.FindOne(ctx, filter)
	err = result.Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
