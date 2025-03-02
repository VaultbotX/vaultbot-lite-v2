package commands

import (
	"context"
	sp "github.com/vaultbotx/vaultbot-lite/internal/spotify"
	"github.com/zmb3/spotify/v2"
)

type SpotifyArtistRepo struct {
	client *sp.Client
}

func (r *SpotifyArtistRepo) GetArtists(artistIds []spotify.ID, artistChan chan<- *spotify.FullArtist, ctx context.Context) error {
	r.client.Mu.Lock()
	defer r.client.Mu.Unlock()

	artists, err := r.client.Client.GetArtists(ctx, artistIds...)
	if err != nil {
		return err
	}

	for _, artist := range artists {
		artistChan <- artist
	}
	close(artistChan)

	return nil
}
