package persistence

import (
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/vaultbotx/vaultbot-lite/internal/domain"
	mg "github.com/vaultbotx/vaultbot-lite/internal/persistence/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Preference struct {
	Id    string               `bson:"_id"`
	Key   domain.PreferenceKey `bson:"key"`
	Value any                  `bson:"value"`
}

type PreferenceRepo struct {
	Client *mongo.Client
}

func (p PreferenceRepo) Set(ctx context.Context, preferenceKey domain.PreferenceKey, value any) error {
	collection := p.Client.Database(mg.DatabaseName).Collection(mg.PreferencesCollection)

	filter := bson.M{"key": preferenceKey}
	update := bson.M{"$set": bson.M{"value": value}}

	upsert := true
	opts := options.UpdateOptions{Upsert: &upsert}
	_, err := collection.UpdateOne(ctx, filter, update, &opts)
	if err != nil {
		return err
	}

	return nil
}

func (p PreferenceRepo) Get(ctx context.Context, preferenceKey domain.PreferenceKey) (*domain.Preference, error) {
	collection := p.Client.Database(mg.DatabaseName).Collection(mg.PreferencesCollection)

	filter := bson.M{"key": preferenceKey}
	var preference Preference
	err := collection.FindOne(ctx, filter).Decode(&preference)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}

	return &domain.Preference{
		Id:    preference.Id,
		Key:   preference.Key,
		Value: preference.Value,
	}, nil
}

func (p PreferenceRepo) GetAll(ctx context.Context) (map[domain.PreferenceKey]domain.Preference, error) {
	collection := p.Client.Database(mg.DatabaseName).Collection(mg.PreferencesCollection)
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

	preferences := make(map[domain.PreferenceKey]domain.Preference)
	for cursor.Next(ctx) {
		var preference Preference
		err2 := cursor.Decode(&preference)
		if err2 != nil {
			return nil, err2
		}
		preferences[preference.Key] = domain.Preference{
			Id:    preference.Id,
			Key:   preference.Key,
			Value: preference.Value,
		}
	}

	if err2 := cursor.Err(); err2 != nil {
		return nil, err2
	}

	return preferences, nil
}
