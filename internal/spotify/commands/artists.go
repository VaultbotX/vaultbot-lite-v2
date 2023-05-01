package commands

import (
	"context"
	sp "github.com/vaultbotx/vaultbot-lite/internal/spotify"
	"github.com/zmb3/spotify/v2"
)

func GetArtists(ctx context.Context, artistIds []spotify.ID, artistChan chan<- *spotify.FullArtist) error {
	client, err := sp.GetSpotifyClient(ctx)
	if err != nil {
		return err
	}

	client.Mu.Lock()
	defer client.Mu.Unlock()

	artists, err := client.Client.GetArtists(ctx, artistIds...)
	if err != nil {
		return err
	}

	for _, artist := range artists {
		artistChan <- artist
	}

	return nil
}
