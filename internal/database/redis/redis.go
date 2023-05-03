package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
	"github.com/vaultbotx/vaultbot-lite/internal/types"
	"os"
	"time"
)

var (
	instance   *redis.Client
	timeFormat = time.RFC3339
)

func getClient() *redis.Client {
	if instance != nil {
		return instance
	}

	host, hostPresent := os.LookupEnv("REDIS_HOST")
	if !hostPresent {
		log.Fatal("Missing REDIS_HOST environment variable")
	}

	password, passwordPresent := os.LookupEnv("REDIS_PASSWORD")
	if !passwordPresent {
		log.Fatal("Missing REDIS_PASSWORD environment variable")
	}

	address := host + ":6379"
	instance = redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
		DB:       0,
	})

	return instance
}

func Flush(ctx context.Context) error {
	return getClient().FlushAll(ctx).Err()
}

func Set(ctx context.Context, track *types.CacheTrack) error {
	addedAt := track.AddedAt.Format(timeFormat)
	return getClient().Set(ctx, track.TrackId, addedAt, 0).Err()
}

func SetMulti(ctx context.Context, tracks []*types.CacheTrack) error {
	trackMap := map[string]string{}
	for _, track := range tracks {
		trackMap[track.TrackId] = track.AddedAt.Format(timeFormat)
	}
	return getClient().MSet(ctx, trackMap).Err()
}

func Get(ctx context.Context, key string) (*string, error) {
	result, err := getClient().Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func GetAll(ctx context.Context, cacheTrackChan chan<- *types.CacheTrack) error {
	var cursor uint64
	client := getClient()
	for {
		var keys []string
		var err error
		keys, cursor, err = client.Scan(ctx, cursor, "*", 10).Result()
		if err != nil {
			return err
		}

		for _, key := range keys {
			hash, err := client.HGetAll(ctx, key).Result()
			if err != nil {
				return err
			}

			for field, value := range hash {
				time, err := time.Parse(timeFormat, value)
				if err != nil {
					return err
				}

				track := types.CacheTrack{
					TrackId: field,
					AddedAt: time,
				}
				cacheTrackChan <- &track
			}
		}

		if cursor == 0 {
			break
		}
	}

	return nil
}
