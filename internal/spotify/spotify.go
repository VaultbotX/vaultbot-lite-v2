package spotify

import (
	"context"
	"os"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/vaultbotx/vaultbot-lite/internal/utils"
	"github.com/zmb3/spotify/v2"
	auth "github.com/zmb3/spotify/v2/auth"
)

var instance *Client

type Client struct {
	DynamicPlaylistId    spotify.ID
	GenrePlaylistId      spotify.ID
	HighScoresPlaylistId spotify.ID
	ThrowbackPlaylistId  spotify.ID
	VarietyPlaylistId    spotify.ID
	Client               *spotify.Client
	Mu                   sync.Mutex
}

func NewSpotifyClient(ctx context.Context) (*Client, error) {
	if instance != nil {
		return instance, nil
	}

	clientId, ok := os.LookupEnv("SPOTIFY_CLIENT_ID")
	if !ok {
		log.Fatal("Missing SPOTIFY_CLIENT_ID environment variable")
	}

	clientSecret, ok := os.LookupEnv("SPOTIFY_CLIENT_SECRET")
	if !ok {
		log.Fatal("Missing SPOTIFY_CLIENT_SECRET environment variable")
	}

	dynamicPlaylistId, ok := os.LookupEnv("SPOTIFY_PLAYLIST_ID")
	if !ok {
		log.Fatal("Missing SPOTIFY_PLAYLIST_ID environment variable")
	}

	genrePlaylistId, ok := os.LookupEnv("GENRE_SPOTIFY_PLAYLIST_ID")
	if !ok {
		log.Fatal("Missing GENRE_SPOTIFY_PLAYLIST_ID environment variable")
	}

	highScoresPlaylistId, ok := os.LookupEnv("HIGH_SCORES_SPOTIFY_PLAYLIST_ID")
	if !ok {
		log.Fatal("Missing HIGH_SCORES_SPOTIFY_PLAYLIST_ID environment variable")
	}

	throwbackPlaylistId, ok := os.LookupEnv("THROWBACK_SPOTIFY_PLAYLIST_ID")
	if !ok {
		log.Fatal("Missing THROWBACK_SPOTIFY_PLAYLIST_ID environment variable")
	}

	varietyPlaylistId, ok := os.LookupEnv("VARIETY_SPOTIFY_PLAYLIST_ID")
	if !ok {
		log.Fatal("Missing VARIETY_SPOTIFY_PLAYLIST_ID environment variable")
	}

	tokenString, ok := os.LookupEnv("SPOTIFY_TOKEN")
	if !ok {
		log.Fatal("Missing SPOTIFY_TOKEN environment variable")
	}

	token, err := utils.ParseTokenString(tokenString)
	if err != nil {
		log.Fatalf("Unable to parse SPOTIFY_TOKEN: %s", err)
	}

	authenticator := auth.New(
		auth.WithClientID(clientId),
		auth.WithClientSecret(clientSecret),
		auth.WithScopes(
			auth.ScopePlaylistModifyPublic,
			auth.ScopePlaylistModifyPrivate,
			auth.ScopePlaylistReadPrivate,
			auth.ScopePlaylistReadCollaborative,
		),
	)

	httpClient := authenticator.Client(context.Background(), token)
	client := spotify.New(httpClient, spotify.WithRetry(true))

	_, err = client.CurrentUser(ctx)
	if err != nil {
		log.Fatalf("Unable to validate Spotify credentials: %v", err)
	}

	instance = &Client{
		DynamicPlaylistId:    spotify.ID(dynamicPlaylistId),
		GenrePlaylistId:      spotify.ID(genrePlaylistId),
		HighScoresPlaylistId: spotify.ID(highScoresPlaylistId),
		ThrowbackPlaylistId:  spotify.ID(throwbackPlaylistId),
		VarietyPlaylistId:    spotify.ID(varietyPlaylistId),
		Client:               client,
		Mu:                   sync.Mutex{},
	}

	return instance, nil
}
