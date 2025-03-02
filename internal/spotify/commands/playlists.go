package commands

import (
	"context"
	"errors"
	sp "github.com/vaultbotx/vaultbot-lite/internal/spotify"
	"github.com/zmb3/spotify/v2"
)

type SpotifyPlaylistRepo struct {
	client *sp.Client
}

// GetPlaylistTracks gets all tracks from the dynamic playlist. It returns them as *spotify.PlaylistItems,
// which includes information about when the track was added to the playlist.
func (r *SpotifyPlaylistRepo) GetPlaylistTracks(playlistItemChan chan<- *spotify.PlaylistItem, ctx context.Context) error {
	r.client.Mu.Lock()
	defer r.client.Mu.Unlock()

	playlistItems, err := r.client.Client.GetPlaylistItems(ctx, r.client.DynamicPlaylistId)
	if err != nil {
		return err
	}

	for _, playlistItem := range playlistItems.Items {
		playlistItemChan <- &playlistItem
	}

	for page := 1; ; page++ {
		err = r.client.Client.NextPage(ctx, playlistItems)
		if errors.Is(err, spotify.ErrNoMorePages) {
			break
		}
		if err != nil {
			return err
		}

		for _, playlistItem := range playlistItems.Items {
			playlistItemChan <- &playlistItem
		}
	}

	close(playlistItemChan)

	return nil
}

func (r *SpotifyPlaylistRepo) AddTracksToPlaylist(ctx context.Context, trackIds []spotify.ID) error {
	r.client.Mu.Lock()
	defer r.client.Mu.Unlock()

	_, err := r.client.Client.AddTracksToPlaylist(ctx, r.client.DynamicPlaylistId, trackIds...)
	if err != nil {
		return err
	}

	return nil
}

func (r *SpotifyPlaylistRepo) RemoveTracksFromPlaylist(ctx context.Context, trackIds []spotify.ID) error {
	r.client.Mu.Lock()
	defer r.client.Mu.Unlock()

	_, err := r.client.Client.RemoveTracksFromPlaylist(ctx, r.client.DynamicPlaylistId, trackIds...)
	if err != nil {
		return err
	}

	return nil
}
