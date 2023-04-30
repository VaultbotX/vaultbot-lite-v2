package spotify

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2/clientcredentials"
	"os"
	"sync"
)

var (
	instance *Client
)

type Client struct {
	client *spotify.Client
	mu     sync.Mutex
}

func GetSpotifyClient() (*Client, error) {
	if instance != nil {
		return instance, nil
	}

	ctx := context.Background()

	clientId, clientIdPresent := os.LookupEnv("SPOTIFY_CLIENT_ID")
	if !clientIdPresent {
		log.Fatal("Missing SPOTIFY_CLIENT_ID environment variable")
	}

	secret, secretPresent := os.LookupEnv("SPOTIFY_CLIENT_SECRET")
	if !secretPresent {
		log.Fatal("Missing SPOTIFY_CLIENT_SECRET environment variable")
	}

	config := &clientcredentials.Config{
		ClientID:     clientId,
		ClientSecret: secret,
		TokenURL:     spotifyauth.TokenURL,
	}
	token, err := config.Token(ctx)
	if err != nil {
		return nil, err
	}

	httpClient := spotifyauth.New().Client(ctx, token)
	client := spotify.New(httpClient)
	instance = &Client{client: client}

	return instance, nil
}
