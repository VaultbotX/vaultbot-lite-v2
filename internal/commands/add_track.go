package commands

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/vaultbotx/vaultbot-lite/internal/database"
	re "github.com/vaultbotx/vaultbot-lite/internal/database/redis"
	sp "github.com/vaultbotx/vaultbot-lite/internal/spotify"
	spcommands "github.com/vaultbotx/vaultbot-lite/internal/spotify/commands"
	"github.com/vaultbotx/vaultbot-lite/internal/types"
	"github.com/zmb3/spotify/v2"
	"sync"
)

func AddTrack(ctx context.Context, trackId string, meta log.Fields) (*spotify.FullTrack, error) {
	log.WithFields(meta).Debugf("Attempting to add track %v to playlist", trackId)
	// 0. Parse the track id
	convertedTrackId := sp.ParseTrackId(trackId)
	if convertedTrackId == nil {
		return nil, types.ErrInvalidTrackId
	}

	// 1. Check the redis cache to see if the track exists
	existingTrack, err := re.Get(ctx, convertedTrackId.String())
	if err != nil {
		log.WithFields(meta).Error(err)
		return nil, err
	}

	if existingTrack != nil {
		log.WithFields(meta).Debugf("Track %v already exists in database", convertedTrackId.String())
		return nil, types.ErrTrackAlreadyInPlaylist
	}

	// 2. Attempt to get the track from Spotify
	trackChan := make(chan *spotify.FullTrack)
	errorChan := make(chan error)
	var wg sync.WaitGroup
	wg.Add(1)

	go func(g *sync.WaitGroup, c chan<- error) {
		defer g.Done()
		err := spcommands.GetTrack(ctx, *convertedTrackId, trackChan)
		if err != nil {
			close(trackChan)
			c <- err
		}
	}(&wg, errorChan)

	wg.Wait()

	var track *spotify.FullTrack
	var ok bool
	select {
	case err := <-errorChan:
		return nil, err
	case track, ok = <-trackChan:
		if track == nil {
			log.WithFields(meta).Debugf("Track %v does not exist", convertedTrackId.String())
			return nil, types.ErrNoTrackExists
		}

		if !ok {
			close(trackChan)
		}
		log.WithFields(meta).Debugf("Track %v exists", convertedTrackId.String())
	}

	log.WithFields(meta).Debugf("Getting artists and audio features for track %v", convertedTrackId.String())
	// 3. If exists, pull the artists and song features
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

		err := spcommands.GetArtists(ctx, artistIds, artistChan)
		if err != nil {
			errorChan <- err
		}
	}(&wg, artistChan)

	go func(wg *sync.WaitGroup, audioFeaturesChan chan<- *spotify.AudioFeatures) {
		defer wg.Done()
		err := spcommands.GetTrackAudioFeatures(ctx, *convertedTrackId, audioFeaturesChan)
		if err != nil {
			errorChan2 <- err
		}
	}(&wg, audioFeaturesChan)

	wg.Wait()

	select {
	case err := <-errorChan:
		close(artistChan)
		log.WithFields(meta).Errorf("Error getting artists: %v", err)
		return nil, err
	case err := <-errorChan2:
		close(audioFeaturesChan)
		log.WithFields(meta).Errorf("Error getting audio features: %v", err)
		return nil, err
	default:
		break
	}

	var artists []*spotify.FullArtist
	for artist := range artistChan {
		artists = append(artists, artist)
	}
	close(artistChan)

	var audioFeatures []*spotify.AudioFeatures
	for audioFeature := range audioFeaturesChan {
		audioFeatures = append(audioFeatures, audioFeature)
	}
	close(audioFeaturesChan)
	log.WithFields(meta).Debugf("Finished getting artists and audio features for track %v", convertedTrackId.String())

	log.WithFields(meta).Debugf("Adding track %v to playlist", convertedTrackId.String())
	// 4. Add to playlist
	err = spcommands.AddTracksToPlaylist(ctx, []spotify.ID{track.ID})
	if err != nil {
		log.WithFields(meta).Errorf("Error adding track to playlist: %v", err)
		return nil, types.ErrCouldNotAddToPlaylist
	}

	log.WithFields(meta).Debugf("Adding track %v to database", convertedTrackId.String())
	// 5. Add to databases
	err = database.AddTrackToDatabase(ctx, track, artists, audioFeatures)
	if err != nil {
		log.WithFields(meta).Errorf("Error adding track to database: %v", err)

		log.WithFields(meta).Debugf("Attempting to rollback adding track %v to playlist", convertedTrackId.String())

		err2 := spcommands.RemoveTracksFromPlaylist(ctx, []spotify.ID{track.ID})
		if err2 != nil {
			log.WithFields(meta).Errorf("Error removing track from playlist: %v", err2)
			return nil, types.ErrCouldNotRemoveFromPlaylist
		}

		return nil, types.ErrCouldNotAddToDatabase
	}

	return track, nil
}
