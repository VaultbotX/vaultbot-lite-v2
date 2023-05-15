package commands

import (
	"context"
	mg "github.com/vaultbotx/vaultbot-lite/internal/database/mongo"
	"github.com/vaultbotx/vaultbot-lite/internal/database/mongo/types"
	internaltypes "github.com/vaultbotx/vaultbot-lite/internal/types"
	"github.com/zmb3/spotify/v2"
	"time"
)

func BlacklistTrack(ctx context.Context, trackId spotify.ID, userFields *internaltypes.UserFields, time time.Time) error {
	instance, err := mg.GetMongoClient(ctx)
	if err != nil {
		return err
	}
	defer instance.Disconnect(ctx)

	collection := instance.Database(mg.DatabaseName).Collection(mg.BlacklistCollection)
	newBlacklistedTrack := types.BlacklistedTrack{
		TrackId:           trackId.String(),
		BlockedById:       userFields.UserId,
		BlockedByUsername: userFields.Username,
		GuildId:           userFields.GuildId,
		Timestamp:         time.Unix(),
	}

	_, err = collection.InsertOne(ctx, newBlacklistedTrack)
	if err != nil {
		return err
	}

	return nil
}

func BlacklistArtist(ctx context.Context, artistId spotify.ID, userFields *internaltypes.UserFields, time time.Time) error {
	instance, err := mg.GetMongoClient(ctx)
	if err != nil {
		return err
	}
	defer instance.Disconnect(ctx)

	collection := instance.Database(mg.DatabaseName).Collection(mg.BlacklistCollection)
	newBlacklistedArtist := types.BlacklistedArtist{
		ArtistId:          artistId.String(),
		BlockedById:       userFields.UserId,
		BlockedByUsername: userFields.Username,
		GuildId:           userFields.GuildId,
		Timestamp:         time.Unix(),
	}

	_, err = collection.InsertOne(ctx, newBlacklistedArtist)
	if err != nil {
		return err
	}

	return nil
}

func BlacklistGenre(ctx context.Context, genreName string, userFields *internaltypes.UserFields, time time.Time) error {
	instance, err := mg.GetMongoClient(ctx)
	if err != nil {
		return err
	}
	defer instance.Disconnect(ctx)

	collection := instance.Database(mg.DatabaseName).Collection(mg.BlacklistCollection)
	newBlacklistedGenre := types.BlacklistedGenre{
		GenreName:         genreName,
		BlockedById:       userFields.UserId,
		BlockedByUsername: userFields.Username,
		GuildId:           userFields.GuildId,
		Timestamp:         time.Unix(),
	}

	_, err = collection.InsertOne(ctx, newBlacklistedGenre)
	if err != nil {
		return err
	}

	return nil
}
