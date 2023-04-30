package commands

import (
	"context"
	log "github.com/sirupsen/logrus"
	sp "github.com/tbrittain/vaultbot-lite/internal/spotify"
	"github.com/zmb3/spotify/v2"
)

// getTrack is a function that is meant to be run as a goroutine to get a track from Spotify
func getTrack(ctx context.Context, id spotify.ID, trackChan chan<- *spotify.FullTrack) error {
	client, err := sp.GetSpotifyClient()
	if err != nil {
		log.Errorf("Error getting Spotify client: %v", err)
		return err
	}

	client.Mu.Lock()
	defer client.Mu.Unlock()

	track, err := client.Client.GetTrack(ctx, id)
	if err != nil {
		log.Errorf("Error getting track: %v", err)
		return err
	}

	trackChan <- track
	close(trackChan)

	return nil
}
