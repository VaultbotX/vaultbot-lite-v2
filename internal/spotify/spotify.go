package spotify

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify/v2"
	auth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2/clientcredentials"
	"os"
	"sync"
)

var (
	instance *Client
)

type Client struct {
	DynamicPlaylistId spotify.ID
	Client            *spotify.Client
	Mu                sync.Mutex
}

func GetSpotifyClient(ctx context.Context) (*Client, error) {
	if instance != nil {
		return instance, nil
	}

	clientId, clientIdPresent := os.LookupEnv("SPOTIFY_CLIENT_ID")
	if !clientIdPresent {
		log.Fatal("Missing SPOTIFY_CLIENT_ID environment variable")
	}

	secret, secretPresent := os.LookupEnv("SPOTIFY_CLIENT_SECRET")
	if !secretPresent {
		log.Fatal("Missing SPOTIFY_CLIENT_SECRET environment variable")
	}

	playlistId, playlistIdPresent := os.LookupEnv("SPOTIFY_PLAYLIST_ID")
	if !playlistIdPresent {
		log.Fatal("Missing SPOTIFY_PLAYLIST_ID environment variable")
	}

	config := &clientcredentials.Config{
		ClientID:     clientId,
		ClientSecret: secret,
		TokenURL:     auth.TokenURL,
		Scopes:       []string{auth.ScopePlaylistModifyPublic},
	}
	token, err := config.Token(ctx)
	if err != nil {
		return nil, err
	}

	httpClient := auth.New().Client(ctx, token)
	client := spotify.New(httpClient)
	instance = &Client{Client: client, DynamicPlaylistId: spotify.ID(playlistId)}

	return instance, nil
}
