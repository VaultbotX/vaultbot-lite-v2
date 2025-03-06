package commands

import (
	"context"
	log "github.com/sirupsen/logrus"
	sp "github.com/vaultbotx/vaultbot-lite/internal/spotify"
	"github.com/zmb3/spotify/v2"
)

type SpotifyTrackRepo struct {
	Client *sp.Client
}

func (r *SpotifyTrackRepo) GetTrack(trackId spotify.ID, trackChan chan<- *spotify.FullTrack, ctx context.Context) error {
	r.Client.Mu.Lock()
	defer r.Client.Mu.Unlock()
	err := r.Client.RefreshAccessTokenIfExpired(ctx)
	if err != nil {
		return err
	}

	track, err := r.Client.Client.GetTrack(ctx, trackId)
	if err != nil {
		log.Errorf("Error getting track: %v", err)
		return err
	}

	trackChan <- track
	close(trackChan)

	return nil
}

// Deprecated: Spotify no longer supports this endpoint
// 1. https://community.spotify.com/t5/Spotify-for-Developers/Changes-to-Web-API/td-p/6540414
// 2. https://developer.spotify.com/blog/2024-11-27-changes-to-the-web-api
func (r *SpotifyTrackRepo) GetTrackAudioFeatures(ctx context.Context, trackId spotify.ID,
	audioFeaturesChan chan<- *spotify.AudioFeatures) error {

	r.Client.Mu.Lock()
	defer r.Client.Mu.Unlock()
	err := r.Client.RefreshAccessTokenIfExpired(ctx)
	if err != nil {
		return err
	}

	audioFeatures, err := r.Client.Client.GetAudioFeatures(ctx, trackId)
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
