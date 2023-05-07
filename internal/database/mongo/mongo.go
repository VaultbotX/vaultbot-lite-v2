package mongo

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
)

var (
	uri *string
)

const (
	DatabaseName          = "vaultbot"
	PreferencesCollection = "preferences"
	AddedTracksCollection = "addedTracks"
)

func GetMongoClient(ctx context.Context) (*mongo.Client, error) {
	if uri == nil {
		// For now, not including any authentication
		host, hostPresent := os.LookupEnv("MONGO_HOST")
		if !hostPresent {
			log.Fatal("Missing MONGO_HOST environment variable")
		}

		combined := fmt.Sprintf("mongodb://%s:27017/?retryWrites=true&w=majority", host)
		uri = &combined
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(*uri))
	if err != nil {
		log.Errorf("Error connecting to MongoDB: %v", err)
		return nil, err
	}

	return client, nil
}
