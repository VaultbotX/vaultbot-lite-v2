package commands

import (
	"context"
	log "github.com/sirupsen/logrus"
	sp "github.com/vaultbotx/vaultbot-lite/internal/spotify"
	"github.com/zmb3/spotify/v2"
)

func GetTrack(ctx context.Context, trackId spotify.ID, trackChan chan<- *spotify.FullTrack) error {
	client, err := sp.GetSpotifyClient(ctx)
	if err != nil {
		log.Errorf("Error getting Spotify client: %v", err)
		return err
	}

	client.Mu.Lock()
	defer client.Mu.Unlock()

	track, err := client.Client.GetTrack(ctx, trackId)
	if err != nil {
		log.Errorf("Error getting track: %v", err)
		return err
	}

	trackChan <- track
	close(trackChan)

	return nil
}

func GetTrackAudioFeatures(ctx context.Context, trackId spotify.ID,
	audioFeaturesChan chan<- *spotify.AudioFeatures) error {
	client, err := sp.GetSpotifyClient(ctx)
	if err != nil {
		log.Errorf("Error getting Spotify client: %v", err)
		return err
	}

	client.Mu.Lock()
	defer client.Mu.Unlock()

	audioFeatures, err := client.Client.GetAudioFeatures(ctx, trackId)
	if err != nil {
		log.Errorf("Error getting audio features: %v", err)
		return err
	}

	// only push one audioFeature, and log if there are multiple since we are only expecting one
	if len(audioFeatures) > 1 {
		log.Warnf("There are multiple audio features for track %v, only using first one", trackId)
	}

	audioFeaturesChan <- audioFeatures[0]
	close(audioFeaturesChan)

	return nil
}
