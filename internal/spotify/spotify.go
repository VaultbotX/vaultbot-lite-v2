package spotify

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/vaultbotx/vaultbot-lite/internal/types"
	"github.com/vaultbotx/vaultbot-lite/internal/utils"
	"github.com/zmb3/spotify/v2"
	auth "github.com/zmb3/spotify/v2/auth"
	"net/http"
	"os"
	"sync"
)

var (
	instance *Client
	state    = "qualified_gopher"
)

const (
	redirectUri = "http://localhost:8080/callback"
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

	clientSecret, clientSecretPresent := os.LookupEnv("SPOTIFY_CLIENT_SECRET")
	if !clientSecretPresent {
		log.Fatal("Missing SPOTIFY_CLIENT_SECRET environment variable")
	}

	playlistId, playlistIdPresent := os.LookupEnv("SPOTIFY_PLAYLIST_ID")
	if !playlistIdPresent {
		log.Fatal("Missing SPOTIFY_PLAYLIST_ID environment variable")
	}

	authenticator := auth.New(
		auth.WithClientID(clientId),
		auth.WithClientSecret(clientSecret),
		auth.WithRedirectURL(redirectUri),
		auth.WithScopes(
			auth.ScopePlaylistModifyPublic,
			auth.ScopePlaylistModifyPrivate,
			auth.ScopePlaylistReadPrivate,
			auth.ScopePlaylistReadCollaborative),
	)

	// https://developer.spotify.com/documentation/web-api/tutorials/code-flow
	// TODO: Here, we will attempt to get an existing token and use it (allowing it to be refreshed if necessary)
	//  If that fails, we will need to open a browser window to get a new token
	// This step will need to occur while running the application locally, and hopefully should only need
	// to happen once

	ch := make(chan *spotify.Client)

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		token, err := authenticator.Token(r.Context(), state, r)
		if err != nil {
			http.Error(w, "Couldn't get token", http.StatusForbidden)
			log.Fatal(err)
		}
		if st := r.FormValue("state"); st != state {
			http.NotFound(w, r)
			log.Fatalf("State mismatch: %s != %s\n", st, state)
		}

		client := spotify.New(authenticator.Client(r.Context(), token))
		log.Info(w, "Successfully retrieved token from Spotify")
		ch <- client
	})

	go func() {
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			log.Fatal(err)
		}
		log.Info("Listening on 8080 for Spotify auth callback")
	}()

	url := authenticator.AuthURL(state)
	err := utils.OpenBrowser(url)
	if err != nil {
		if err == types.ErrUnsupportedOSForBrowser {
			log.Warnf("Unable to automatically open browser. Please log in to Spotify by visiting "+
				"the following page in your browser: %s", url)
		} else {
			log.Fatal(err)
		}
	}

	log.Info("Waiting for Spotify auth callback")
	client := <-ch

	_, err = client.CurrentUser(ctx)
	if err != nil {
		log.Fatalf("Unable to get current user. This application requires user-level permissions to perform"+
			"various playlist operations: %v", err)
	}

	instance = &Client{
		DynamicPlaylistId: spotify.ID(playlistId),
		Client:            client,
	}

	return instance, nil
}
