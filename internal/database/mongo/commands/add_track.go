package commands

import (
	"context"
	"github.com/vaultbotx/vaultbot-lite/internal/database/mongo"
	"github.com/vaultbotx/vaultbot-lite/internal/database/mongo/types"
	internaltypes "github.com/vaultbotx/vaultbot-lite/internal/types"
	"github.com/zmb3/spotify/v2"
	"time"
)

func AddTrack(ctx context.Context, trackId spotify.ID, userFields *internaltypes.UserFields, time time.Time) error {
	instance, err := mongo.GetMongoClient(ctx)
	if err != nil {
		return err
	}
	defer instance.Disconnect(ctx)

	collection := instance.Database(mongo.DatabaseName).Collection(mongo.AddedTracksCollection)
	newAddedTrack := types.AddedTrack{
		TrackId:   trackId.String(),
		UserId:    userFields.UserId,
		Username:  userFields.Username,
		GuildId:   userFields.GuildId,
		Timestamp: time.Unix(),
	}

	_, err = collection.InsertOne(ctx, newAddedTrack)
	if err != nil {
		return err
	}

	return nil
}
