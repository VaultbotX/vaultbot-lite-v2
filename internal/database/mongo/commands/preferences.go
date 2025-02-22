package commands

import (
	"context"
	"errors"
	mg "github.com/vaultbotx/vaultbot-lite/internal/database/mongo"
	"github.com/vaultbotx/vaultbot-lite/internal/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func SetPreference(ctx context.Context, key types.PreferenceKey, value interface{}) error {
	instance, err := mg.GetMongoClient(ctx)
	if err != nil {
		return err
	}
	defer instance.Disconnect(ctx)

	collection := instance.Database(mg.DatabaseName).Collection(mg.PreferencesCollection)

	filter := bson.M{"key": key}
	update := bson.M{"$set": bson.M{"value": value}}

	upsert := true
	opts := options.UpdateOptions{Upsert: &upsert}
	_, err = collection.UpdateOne(context.Background(), filter, update, &opts)
	if err != nil {
		return err
	}

	return nil
}

func GetPreference(ctx context.Context, key types.PreferenceKey) (*types.Preference, error) {
	instance, err := mg.GetMongoClient(ctx)
	if err != nil {
		return nil, err
	}
	defer instance.Disconnect(ctx)

	collection := instance.Database(mg.DatabaseName).Collection(mg.PreferencesCollection)

	filter := bson.M{"key": key}
	var preference types.Preference
	err = collection.FindOne(ctx, filter).Decode(&preference)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}

	return &preference, nil
}

func GetAllPreferences(ctx context.Context) (map[types.PreferenceKey]types.Preference, error) {
	instance, err := mg.GetMongoClient(ctx)
	if err != nil {
		return nil, err
	}
	defer instance.Disconnect(ctx)

	collection := instance.Database(mg.DatabaseName).Collection(mg.PreferencesCollection)
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	preferences := make(map[types.PreferenceKey]types.Preference)
	for cursor.Next(ctx) {
		var preference types.Preference
		err2 := cursor.Decode(&preference)
		if err2 != nil {
			return nil, err2
		}
		preferences[preference.Key] = preference
	}

	if err2 := cursor.Err(); err2 != nil {
		return nil, err2
	}

	return preferences, nil
}
