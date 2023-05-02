package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
	"os"
)

var (
	instance *redis.Client
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

func Set(ctx context.Context, key string, value string) error {
	return getClient().Set(ctx, key, value, 0).Err()
}

func SetMulti(ctx context.Context, values map[string]string) error {
	return getClient().MSet(ctx, values).Err()
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
