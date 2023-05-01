package commands

import (
	"context"
	"github.com/tbrittain/vaultbot-lite/internal/spotify/commands"
	"github.com/tbrittain/vaultbot-lite/internal/types"
	"github.com/zmb3/spotify/v2"
	"sync"
)

func AddTrack(ctx context.Context, trackId string) error {
	// 1. Attempt to get the track from Spotify
	convertedTrackId := spotify.ID(trackId)

	trackChan := make(chan *spotify.FullTrack)
	errorChan := make(chan error)
	var wg sync.WaitGroup
	wg.Add(1)

	go func(g *sync.WaitGroup, c chan<- error) {
		defer g.Done()
		err := commands.GetTrack(ctx, convertedTrackId, trackChan)
		if err != nil {
			close(trackChan)
			c <- err
		}
	}(&wg, errorChan)

	var track *spotify.FullTrack
	var ok bool

	wg.Wait()

	err := <-errorChan
	if err != nil {
		return err
	}

	track = <-trackChan
	close(trackChan)
	if track == nil {
		return types.ErrNoTrackExists
	}

	select {
	case err := <-errorChan:
		return err
	case track, ok = <-trackChan:
		if track == nil {
			return types.ErrNoTrackExists
		}

		if !ok {
			close(trackChan)
		}
	}

	// 2. If exists, pull the artists and song features
	artistChan := make(chan *spotify.FullArtist)
	audioFeaturesChan := make(chan *spotify.AudioFeatures)
	errorChan2 := make(chan error)

	wg.Add(2)
	go func(wg *sync.WaitGroup, artistChan chan<- *spotify.FullArtist) {
		defer wg.Done()
		artistIds := make([]spotify.ID, len(track.Artists))
		for i, artist := range track.Artists {
			artistIds[i] = artist.ID
		}

		err := commands.GetArtists(ctx, artistIds, artistChan)
		if err != nil {
			close(artistChan)
			errorChan <- err
		}
	}(&wg, artistChan)

	go func(wg *sync.WaitGroup, audioFeaturesChan chan<- *spotify.AudioFeatures) {
		defer wg.Done()
		err := commands.GetTrackAudioFeatures(ctx, convertedTrackId, audioFeaturesChan)
		if err != nil {
			close(audioFeaturesChan)
			errorChan2 <- err
		}
	}(&wg, audioFeaturesChan)

	wg.Wait()

	select {
	case err := <-errorChan:
		return err
	case err := <-errorChan2:
		return err
	default:
		break
	}

	var artists []*spotify.FullArtist
	for artist := range artistChan {
		artists = append(artists, artist)
	}

	var audioFeatures []*spotify.AudioFeatures
	for audioFeature := range audioFeaturesChan {
		audioFeatures = append(audioFeatures, audioFeature)
	}

	// 3. Add to playlist
	// commands.AddTracksToPlaylist()

	// 4. Add to databases

	// If failure during step 4, rollback adding to playlist

	return nil
}
