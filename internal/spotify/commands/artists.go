package commands

import (
	"context"
	sp "github.com/vaultbotx/vaultbot-lite/internal/spotify"
	"github.com/zmb3/spotify/v2"
)

type SpotifyArtistRepo struct {
	Client *sp.Client
}

func (r *SpotifyArtistRepo) GetArtists(artistIds []spotify.ID, artistChan chan<- *spotify.FullArtist, ctx context.Context) error {
	r.Client.Mu.Lock()
	defer r.Client.Mu.Unlock()
	err := r.Client.RefreshAccessTokenIfExpired(ctx)
	if err != nil {
		return err
	}

	artists, err := r.Client.Client.GetArtists(ctx, artistIds...)
	if err != nil {
		return err
	}

	for _, artist := range artists {
		artistChan <- artist
	}
	close(artistChan)

	return nil
}
