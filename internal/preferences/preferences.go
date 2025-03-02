package preferences

import (
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	mg "github.com/vaultbotx/vaultbot-lite/internal/persistence/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PreferenceKey string

const (
	MaxDurationKey    PreferenceKey = "maxDuration"
	PurgeFrequencyKey PreferenceKey = "purgeFrequency"
	MaxTrackAgeKey    PreferenceKey = "maxTrackAge"
)

type Preference struct {
	Id    string        `bson:"_id"`
	Key   PreferenceKey `bson:"key"`
	Value interface{}   `bson:"value"`
}

var AllPreferences = [3]PreferenceKey{
	MaxDurationKey,
	PurgeFrequencyKey,
	MaxTrackAgeKey,
}

func (key PreferenceKey) DefaultValue() interface{} {
	switch key {
	case MaxDurationKey:
		// 10 minutes in MS
		return 10 * 60 * 1000
	case PurgeFrequencyKey:
		// 12 hours in MS
		return 12 * 60 * 60 * 1000
	case MaxTrackAgeKey:
		// 2 weeks in MS
		return 2 * 7 * 24 * 60 * 60 * 1000
	default:
		return nil
	}
}

func SetPreference(instance *mongo.Client, key PreferenceKey, value interface{}, ctx context.Context) error {
	defer func(instance *mongo.Client, ctx context.Context) {
		err := instance.Disconnect(ctx)
		if err != nil {
			log.Errorf("Error disconnecting from MongoDB: %s", err)
			return
		}
	}(instance, ctx)

	collection := instance.Database(mg.DatabaseName).Collection(mg.PreferencesCollection)

	filter := bson.M{"key": key}
	update := bson.M{"$set": bson.M{"value": value}}

	upsert := true
	opts := options.UpdateOptions{Upsert: &upsert}
	_, err := collection.UpdateOne(ctx, filter, update, &opts)
	if err != nil {
		return err
	}

	return nil
}

func GetPreference(ctx context.Context, key PreferenceKey) (*Preference, error) {
	instance, err := mg.GetMongoClient(ctx)
	if err != nil {
		return nil, err
	}
	defer func(instance *mongo.Client, ctx context.Context) {
		err := instance.Disconnect(ctx)
		if err != nil {
			log.Errorf("Error disconnecting from MongoDB: %s", err)
			return
		}
	}(instance, ctx)

	collection := instance.Database(mg.DatabaseName).Collection(mg.PreferencesCollection)

	filter := bson.M{"key": key}
	var preference Preference
	err = collection.FindOne(ctx, filter).Decode(&preference)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}

	return &preference, nil
}

func GetAllPreferences(ctx context.Context) (map[PreferenceKey]Preference, error) {
	instance, err := mg.GetMongoClient(ctx)
	if err != nil {
		return nil, err
	}
	defer func(instance *mongo.Client, ctx context.Context) {
		err := instance.Disconnect(ctx)
		if err != nil {
			log.Errorf("Error disconnecting from MongoDB: %s", err)
			return
		}
	}(instance, ctx)

	collection := instance.Database(mg.DatabaseName).Collection(mg.PreferencesCollection)
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			log.Errorf("Error closing MongoDB cursor: %s", err)
			return
		}
	}(cursor, ctx)

	preferences := make(map[PreferenceKey]Preference)
	for cursor.Next(ctx) {
		var preference Preference
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
