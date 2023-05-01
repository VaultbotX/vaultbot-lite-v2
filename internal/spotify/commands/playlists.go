package commands

import (
	"context"
	sp "github.com/vaultbotx/vaultbot-lite/internal/spotify"
	"github.com/zmb3/spotify/v2"
)

func GetPlaylistTracks(ctx context.Context, trackChan chan<- *spotify.FullTrack) error {
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
		trackChan <- playlistItem.Track.Track
	}

	for page := 1; ; page++ {
		err = client.Client.NextPage(ctx, playlistItems)
		if err == spotify.ErrNoMorePages {
			break
		}
		if err != nil {
			return err
		}

		for _, playlistItem := range playlistItems.Items {
			trackChan <- playlistItem.Track.Track
		}
	}

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
