package persistence

import (
	"context"
	"errors"
	"github.com/vaultbotx/vaultbot-lite/internal/domain"
	mg "github.com/vaultbotx/vaultbot-lite/internal/persistence/mongo"
	"github.com/vaultbotx/vaultbot-lite/internal/persistence/mongo/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type BlacklistRepository struct {
	client *mongo.Client
}

func NewBlacklistRepository(client *mongo.Client) *BlacklistRepository {
	return &BlacklistRepository{
		client: client,
	}
}

func (r *BlacklistRepository) AddToBlacklist(ctx context.Context, blacklistType domain.BlacklistType, id string,
	userFields *domain.UserFields, time time.Time) error {

	collection := r.client.Database(mg.DatabaseName).Collection(mg.BlacklistCollection)
	var blacklistedItem interface{}
	switch blacklistType {
	case domain.Track:
		blacklistedItem = types.BlacklistedTrack{
			TrackId:           id,
			BlockedById:       userFields.UserId,
			BlockedByUsername: userFields.Username,
			CommonFields: types.CommonFields{
				GuildId:   userFields.GuildId,
				Timestamp: time.Unix(),
			},
		}
	case domain.Artist:
		blacklistedItem = types.BlacklistedArtist{
			ArtistId:          id,
			BlockedById:       userFields.UserId,
			BlockedByUsername: userFields.Username,
			CommonFields: types.CommonFields{
				GuildId:   userFields.GuildId,
				Timestamp: time.Unix(),
			},
		}
	case domain.Genre:
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

	_, err := collection.InsertOne(ctx, blacklistedItem)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return domain.ErrBlacklistItemAlreadyExists
		}
		return err
	}

	return nil
}

func (r *BlacklistRepository) RemoveFromBlacklist(ctx context.Context, blacklistType domain.BlacklistType, id string) error {
	collection := r.client.Database(mg.DatabaseName).Collection(mg.BlacklistCollection)

	var filter bson.M
	switch blacklistType {
	case domain.Track:
		filter = bson.M{"trackId": id}
	case domain.Artist:
		filter = bson.M{"artistId": id}
	case domain.Genre:
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

func (r *BlacklistRepository) CheckBlacklistItem(ctx context.Context, blacklistType domain.BlacklistType, id string) (bool, error) {
	collection := r.client.Database(mg.DatabaseName).Collection(mg.BlacklistCollection)

	var filter bson.M
	switch blacklistType {
	case domain.Track:
		filter = bson.M{"trackId": id}
	case domain.Artist:
		filter = bson.M{"artistId": id}
	case domain.Genre:
		filter = bson.M{"genreName": id}
	}

	result := collection.FindOne(ctx, filter)
	err := result.Err()
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
