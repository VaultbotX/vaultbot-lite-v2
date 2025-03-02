package commands

import (
	"context"
	"errors"
	sp "github.com/vaultbotx/vaultbot-lite/internal/spotify"
	"github.com/zmb3/spotify/v2"
)

// GetPlaylistTracks gets all tracks from the dynamic playlist. It returns them as *spotify.PlaylistItems,
// which includes information about when the track was added to the playlist.
func GetPlaylistTracks(playlistItemChan chan<- *spotify.PlaylistItem, ctx context.Context) error {
	client, err := sp.GetSpotifyClient(ctx)
	if err != nil {
		return err
	}

	client.Mu.Lock()
	defer client.Mu.Unlock()

	playlistItems, err := client.Client.GetPlaylistItems(ctx, client.DynamicPlaylistId)
	if err != nil {
		return err
	}

	for _, playlistItem := range playlistItems.Items {
		playlistItemChan <- &playlistItem
	}

	for page := 1; ; page++ {
		err = client.Client.NextPage(ctx, playlistItems)
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

func AddTracksToPlaylist(ctx context.Context, trackIds []spotify.ID) error {
	client, err := sp.GetSpotifyClient(ctx)
	if err != nil {
		return err
	}

	client.Mu.Lock()
	defer client.Mu.Unlock()

	_, err = client.Client.AddTracksToPlaylist(ctx, client.DynamicPlaylistId, trackIds...)
	if err != nil {
		return err
	}

	return nil
}

func RemoveTracksFromPlaylist(ctx context.Context, trackIds []spotify.ID) error {
	client, err := sp.GetSpotifyClient(ctx)
	if err != nil {
		return err
	}

	client.Mu.Lock()
	defer client.Mu.Unlock()

	_, err = client.Client.RemoveTracksFromPlaylist(ctx, client.DynamicPlaylistId, trackIds...)
	if err != nil {
		return err
	}

	return nil
}
