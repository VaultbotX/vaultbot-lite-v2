package commands

import (
	"context"
	sp "github.com/tbrittain/vaultbot-lite/internal/spotify"
	"github.com/zmb3/spotify/v2"
)

func getPlaylistTracks(ctx context.Context, playlistId spotify.ID, trackChan chan<- *spotify.FullTrack) error {
	client, err := sp.GetSpotifyClient(ctx)
	if err != nil {
		return err
	}

	client.Mu.Lock()
	defer client.Mu.Unlock()

	playlistItems, err := client.Client.GetPlaylistItems(ctx, playlistId)
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

func addTracksToPlaylist(ctx context.Context, playlistId spotify.ID, trackIds []spotify.ID) error {
	client, err := sp.GetSpotifyClient(ctx)
	if err != nil {
		return err
	}

	client.Mu.Lock()
	defer client.Mu.Unlock()

	_, err = client.Client.AddTracksToPlaylist(ctx, playlistId, trackIds...)
	if err != nil {
		return err
	}

	return nil
}
