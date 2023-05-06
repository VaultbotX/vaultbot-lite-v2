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

	spotifyTokenString, spotifyTokenStringPresent := os.LookupEnv("SPOTIFY_TOKEN")
	if spotifyTokenStringPresent {
		// https://developer.spotify.com/documentation/web-api/tutorials/code-flow
		// TODO: Here, we will attempt to get an existing token and use it (allowing it to be refreshed if necessary)
		//  If that fails, we will need to open a browser window to get a new token
		// This step will need to occur while running the application locally, and hopefully should only need
		// to happen once
		token, err := utils.ParseTokenString(spotifyTokenString)
		if err != nil {
			log.Fatalf("Unable to parse Spotify token from environment variable: %s", err)
		}

		client := spotify.New(authenticator.Client(ctx, token))

		validateUserPresent(ctx, client)

		instance = &Client{
			DynamicPlaylistId: spotify.ID(playlistId),
			Client:            client,
		}

		return instance, nil
	}

	// At this point, we were not provided an existing token, so we will need to open a browser window to get one
	// This step will need to occur while running the application locally, and hopefully should only need
	// to happen once
	state, err := utils.GenerateState()
	if err != nil {
		log.Fatalf("Unable to generate state for Spotify auth: %s", err)
	}
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
	err = utils.OpenBrowser(url)
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

	validateUserPresent(ctx, client)

	token, err := client.Token()
	if err != nil {
		log.Fatalf("Unable to get token after completing Spotify client setup. This should not happen: %v", err)
	}
	// write the token to a text file
	err = utils.WriteTokenToFile(token)
	if err != nil {
		log.Fatalf("Unable to write token to file: %v", err)
	}
	log.Warn("Token written to file. Please set the SPOTIFY_TOKEN environment variable to the contents of the file")

	instance = &Client{
		DynamicPlaylistId: spotify.ID(playlistId),
		Client:            client,
	}

	return instance, nil
}

func validateUserPresent(ctx context.Context, client *spotify.Client) {
	_, err := client.CurrentUser(ctx)
	if err != nil {
		log.Fatalf("Unable to get current user. This application requires user-level permissions to perform"+
			"various playlist operations: %v", err)
	}
}
