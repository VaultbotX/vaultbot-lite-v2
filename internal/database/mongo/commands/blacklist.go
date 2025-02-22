package commands

import (
	"context"
	"errors"
	mg "github.com/vaultbotx/vaultbot-lite/internal/database/mongo"
	"github.com/vaultbotx/vaultbot-lite/internal/database/mongo/types"
	internaltypes "github.com/vaultbotx/vaultbot-lite/internal/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

func Blacklist(ctx context.Context, blacklistType internaltypes.BlacklistType, id string,
	userFields *internaltypes.UserFields, time time.Time) error {
	instance, err := mg.GetMongoClient(ctx)
	if err != nil {
		return err
	}
	defer instance.Disconnect(ctx)

	collection := instance.Database(mg.DatabaseName).Collection(mg.BlacklistCollection)
	var blacklistedItem interface{}
	switch blacklistType {
	case internaltypes.Track:
		blacklistedItem = types.BlacklistedTrack{
			TrackId:           id,
			BlockedById:       userFields.UserId,
			BlockedByUsername: userFields.Username,
			CommonFields: types.CommonFields{
				GuildId:   userFields.GuildId,
				Timestamp: time.Unix(),
			},
		}
	case internaltypes.Artist:
		blacklistedItem = types.BlacklistedArtist{
			ArtistId:          id,
			BlockedById:       userFields.UserId,
			BlockedByUsername: userFields.Username,
			CommonFields: types.CommonFields{
				GuildId:   userFields.GuildId,
				Timestamp: time.Unix(),
			},
		}
	case internaltypes.Genre:
		blacklistedItem = types.BlacklistedGenre{
			GenreName:         id,
			BlockedById:       userFields.UserId,
			BlockedByUsername: userFields.Username,
			CommonFields: types.CommonFields{
				GuildId:   userFields.GuildId,
				Timestamp: time.Unix(),
			},
		}
	}

	_, err = collection.InsertOne(ctx, blacklistedItem)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return internaltypes.ErrBlacklistItemAlreadyExists
		}
		return err
	}

	return nil
}

func Unblacklist(ctx context.Context, blacklistType internaltypes.BlacklistType, id string,
	userFields *internaltypes.UserFields) error {
	instance, err := mg.GetMongoClient(ctx)
	if err != nil {
		return err
	}
	defer instance.Disconnect(ctx)

	collection := instance.Database(mg.DatabaseName).Collection(mg.BlacklistCollection)

	var filter bson.M
	switch blacklistType {
	case internaltypes.Track:
		filter = bson.M{"trackId": id}
	case internaltypes.Artist:
		filter = bson.M{"artistId": id}
	case internaltypes.Genre:
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

func CheckBlacklistItem(ctx context.Context, blacklistType internaltypes.BlacklistType, id string) (bool, error) {
	instance, err := mg.GetMongoClient(ctx)
	if err != nil {
		return false, err
	}
	defer instance.Disconnect(ctx)

	collection := instance.Database(mg.DatabaseName).Collection(mg.BlacklistCollection)

	var filter bson.M
	switch blacklistType {
	case internaltypes.Track:
		filter = bson.M{"trackId": id}
	case internaltypes.Artist:
		filter = bson.M{"artistId": id}
	case internaltypes.Genre:
		filter = bson.M{"genreName": id}
	}

	result := collection.FindOne(ctx, filter)
	err = result.Err()
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
