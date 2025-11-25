package commands

import (
	"context"
	"errors"

	sp "github.com/vaultbotx/vaultbot-lite/internal/spotify"
	"github.com/zmb3/spotify/v2"
)

type SpotifyPlaylistRepo struct {
	Client *sp.Client
}

// GetPlaylistTracks gets all tracks from the dynamic playlist. It returns them as *spotify.PlaylistItems,
// which includes information about when the track was added to the playlist.
func (r *SpotifyPlaylistRepo) GetPlaylistTracks(ctx context.Context) ([]spotify.PlaylistItem, error) {
	r.Client.Mu.Lock()
	defer r.Client.Mu.Unlock()

	var playlistItems []spotify.PlaylistItem
	playlistItemsCollection, err := r.Client.Client.GetPlaylistItems(ctx, r.Client.DynamicPlaylistId)
	if err != nil {
		return nil, err
	}

	for _, playlistItem := range playlistItemsCollection.Items {
		playlistItems = append(playlistItems, playlistItem)
	}

	for page := 1; ; page++ {
		err = r.Client.Client.NextPage(ctx, playlistItemsCollection)
		if errors.Is(err, spotify.ErrNoMorePages) {
			break
		}
		if err != nil {
			return nil, err
		}

		for _, playlistItem := range playlistItemsCollection.Items {
			playlistItems = append(playlistItems, playlistItem)
		}
	}

	return playlistItems, nil
}

func (r *SpotifyPlaylistRepo) AddTracksToPlaylist(ctx context.Context, trackIds []spotify.ID) error {
	r.Client.Mu.Lock()
	defer r.Client.Mu.Unlock()

	_, err := r.Client.Client.AddTracksToPlaylist(ctx, r.Client.DynamicPlaylistId, trackIds...)
	if err != nil {
		return err
	}

	return nil
}

func (r *SpotifyPlaylistRepo) RemoveTracksFromPlaylist(ctx context.Context, trackIds []spotify.ID) error {
	r.Client.Mu.Lock()
	defer r.Client.Mu.Unlock()

	_, err := r.Client.Client.RemoveTracksFromPlaylist(ctx, r.Client.DynamicPlaylistId, trackIds...)
	if err != nil {
		return err
	}

	return nil
}
