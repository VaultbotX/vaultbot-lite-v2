package commands

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/vaultbotx/vaultbot-lite/internal/database"
	mongocommands "github.com/vaultbotx/vaultbot-lite/internal/database/mongo/commands"
	sp "github.com/vaultbotx/vaultbot-lite/internal/spotify"
	spcommands "github.com/vaultbotx/vaultbot-lite/internal/spotify/commands"
	"github.com/vaultbotx/vaultbot-lite/internal/types"
	"github.com/zmb3/spotify/v2"
)

func AddTrack(ctx context.Context, trackId string, userFields *types.UserFields, meta log.Fields) (*spotify.FullTrack, error) {
	log.WithFields(meta).Debugf("Attempting to add track %v to playlist", trackId)
	// 0. Parse the track id
	convertedTrackId := sp.ParseTrackId(trackId)
	if convertedTrackId == nil {
		return nil, types.ErrInvalidTrackId
	}

	// 1. Check the cache to see if the track exists
	existingTrack := database.Cache.Get(*convertedTrackId)
	if existingTrack != nil {
		log.WithFields(meta).Debugf("Track %v already exists in database", convertedTrackId.String())
		return nil, types.ErrTrackAlreadyInPlaylist
	}

	// 2. Attempt to get the track from Spotify
	trackChan := make(chan *spotify.FullTrack)
	errorChan := make(chan error)

	go func(c chan<- error) {
		err := spcommands.GetTrack(ctx, *convertedTrackId, trackChan)
		if err != nil {
			close(trackChan)
			c <- err
		}
	}(errorChan)

	var track *spotify.FullTrack
	done := false
	for !done {
		select {
		case err := <-errorChan:
			return nil, err
		case track = <-trackChan:
			if track == nil {
				log.WithFields(meta).Debugf("Track %v does not exist", convertedTrackId.String())
				return nil, types.ErrNoTrackExists
			}

			done = true
			log.WithFields(meta).Debugf("Track %v exists", convertedTrackId.String())
		}
	}

	err := handleTrackOrArtistBlacklisted(ctx, track, meta)
	if err != nil {
		return nil, err
	}

	err = handleMaxDuration(err, track, meta, convertedTrackId)
	if err != nil {
		return nil, err
	}

	log.WithFields(meta).Debugf("Getting artists and audio features for track %v", convertedTrackId.String())
	// 3. If exists, pull the artists and song features
	artistChan := make(chan *spotify.FullArtist)
	audioFeaturesChan := make(chan *spotify.AudioFeatures)
	errorChan2 := make(chan error)

	go func(artistChan chan<- *spotify.FullArtist) {
		artistIds := make([]spotify.ID, len(track.Artists))
		for i, artist := range track.Artists {
			artistIds[i] = artist.ID
		}

		err := spcommands.GetArtists(ctx, artistIds, artistChan)
		if err != nil {
			errorChan <- err
		}
	}(artistChan)

	go func(audioFeaturesChan chan<- *spotify.AudioFeatures) {
		err := spcommands.GetTrackAudioFeatures(ctx, *convertedTrackId, audioFeaturesChan)
		if err != nil {
			errorChan2 <- err
		}
	}(audioFeaturesChan)

	var artists []*spotify.FullArtist
	var audioFeatures []*spotify.AudioFeatures
	artistsDone, audioFeaturesDone := false, false
	for !artistsDone || !audioFeaturesDone {
		select {
		case err := <-errorChan:
			close(artistChan)
			log.WithFields(meta).Errorf("Error getting artists: %v", err)
			return nil, err
		case err := <-errorChan2:
			close(audioFeaturesChan)
			log.WithFields(meta).Errorf("Error getting audio features: %v", err)
			return nil, err
		case artist, ok := <-artistChan:
			if artist != nil {
				artists = append(artists, artist)
			}

			if !ok {
				artistsDone = true
			}
		case audioFeature, ok := <-audioFeaturesChan:
			if audioFeature != nil {
				audioFeatures = append(audioFeatures, audioFeature)
			}

			if !ok {
				audioFeaturesDone = true
			}
		}
	}

	log.WithFields(meta).Debugf("Finished getting artists and audio features for track %v", convertedTrackId.String())
	// 2.3 Check each of the genres for each artist and ensure that none of them are blacklisted
	for _, artist := range artists {
		for _, genre := range artist.Genres {
			genreBlacklisted, err := mongocommands.CheckBlacklistItem(ctx, mongocommands.Genre, genre)
			if err != nil {
				return nil, err
			}

			if genreBlacklisted {
				log.WithFields(meta).Debugf("Genre %v for artist %v is blacklisted", genre, artist.Name)
				return nil, &types.ErrGenreBlacklisted{GenreName: genre, ArtistName: artist.Name}
			}
		}
	}

	log.WithFields(meta).Debugf("Adding track %v to playlist", convertedTrackId.String())
	// 4. Add to playlist
	err = spcommands.AddTracksToPlaylist(ctx, []spotify.ID{track.ID})
	if err != nil {
		log.WithFields(meta).Errorf("Error adding track to playlist: %v", err)
		return nil, types.ErrCouldNotAddToPlaylist
	}

	log.WithFields(meta).Debugf("Adding track %v to database", convertedTrackId.String())
	// 5. Add to databases
	err = database.AddTrackToDatabase(ctx, userFields, track, artists, audioFeatures)
	if err != nil {
		log.WithFields(meta).Errorf("Error adding track to database: %v", err)

		log.WithFields(meta).Debugf("Attempting to rollback adding track %v to playlist", convertedTrackId.String())

		err2 := spcommands.RemoveTracksFromPlaylist(ctx, []spotify.ID{track.ID})
		if err2 != nil {
			log.WithFields(meta).Errorf("Error removing track from playlist during rollback: %v", err2)
			return nil, types.ErrCouldNotRemoveFromPlaylist
		}

		return nil, types.ErrCouldNotAddToDatabase
	}

	return track, nil
}

func handleMaxDuration(err error, track *spotify.FullTrack, meta log.Fields, convertedTrackId *spotify.ID) error {
	// 2.5 Check that the duration of the song does not exceed the maximum
	maxDurationPreference, err := GetMaxSongDurationPreference()
	if err != nil {
		return err
	}
	var maxDuration int
	if v, ok := maxDurationPreference.Value.(int32); ok {
		maxDuration = int(v)
	} else {
		log.Warn("Max duration preference is not an int32, using default value")
		maxDuration = types.MaxDurationKey.DefaultValue().(int)
	}

	if track.Duration > maxDuration {
		log.WithFields(meta).Debugf("Track %v exceeds maximum duration", convertedTrackId.String())
		return types.ErrTrackTooLong
	}

	return nil
}

func handleTrackOrArtistBlacklisted(ctx context.Context, track *spotify.FullTrack, meta log.Fields) error {
	// 2.1 Check that the track is not blacklisted
	trackBlacklisted, err := mongocommands.CheckBlacklistItem(ctx, mongocommands.Track, track.ID.String())
	if err != nil {
		return err
	}

	if trackBlacklisted {
		log.WithFields(meta).Debugf("Track %v is blacklisted", track.Name)
		artistNames := make([]string, len(track.Artists))
		for i, artist := range track.Artists {
			artistNames[i] = artist.Name
		}
		return &types.ErrTrackBlacklisted{TrackName: track.Name, ArtistNames: artistNames}
	}

	// 2.2 Check each of the artists and ensure that none of them are blacklisted
	for _, artist := range track.Artists {
		artistBlacklisted, err := mongocommands.CheckBlacklistItem(ctx, mongocommands.Artist, artist.ID.String())
		if err != nil {
			return err
		}

		if artistBlacklisted {
			log.WithFields(meta).Debugf("Artist %v is blacklisted", artist.ID.String())
			return &types.ErrArtistBlacklisted{ArtistName: artist.Name}
		}
	}

	return nil
}
