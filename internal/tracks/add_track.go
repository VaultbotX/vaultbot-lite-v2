package tracks

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/vaultbotx/vaultbot-lite/internal/domain"
	"github.com/vaultbotx/vaultbot-lite/internal/persistence"
	sp "github.com/vaultbotx/vaultbot-lite/internal/spotify"
	"github.com/zmb3/spotify/v2"
)

type AddTrackInput struct {
	TrackId    string
	UserFields *domain.UserFields
	Ctx        context.Context
	Meta       log.Fields

	TrackService     *domain.TrackService
	BlacklistService *domain.BlacklistService

	SpTrackService    *domain.SpotifyTrackService
	SpArtistService   *domain.SpotifyArtistService
	SpPlaylistService *domain.SpotifyPlaylistService

	PreferenceService *domain.PreferenceService
}

func AddTrack(input *AddTrackInput) (*spotify.FullTrack, error) {
	log.WithFields(input.Meta).Debugf("Attempting to add track %v to playlist", input.TrackId)
	// 0. Parse the track id
	convertedTrackId := sp.ParseTrackId(input.TrackId)
	if convertedTrackId == nil {
		return nil, domain.ErrInvalidTrackId
	}

	// 1. Check the cache to see if the track exists
	existingTrack := persistence.TrackCache.Get(*convertedTrackId)
	if existingTrack != nil {
		log.WithFields(input.Meta).Debugf("Track %v already exists in database", convertedTrackId.String())
		return nil, domain.ErrTrackAlreadyInPlaylist
	}

	// 2. Attempt to get the track from Spotify
	trackChan := make(chan *spotify.FullTrack)
	errorChan := make(chan error)

	go func(c chan<- error) {
		err := input.SpTrackService.Repo.GetTrack(*convertedTrackId, trackChan, input.Ctx)
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
				log.WithFields(input.Meta).Debugf("Track %v does not exist", convertedTrackId.String())
				return nil, domain.ErrNoTrackExists
			}

			done = true
			log.WithFields(input.Meta).Debugf("Track %v exists", convertedTrackId.String())
		}
	}

	err := handleTrackOrArtistBlacklisted(input.BlacklistService, track, input.Ctx, input.Meta)
	if err != nil {
		return nil, err
	}

	err = handleMaxDuration(input.PreferenceService, track, input.Meta, convertedTrackId, input.Ctx)
	if err != nil {
		return nil, err
	}

	log.WithFields(input.Meta).Debugf("Getting artists and audio features for track %v", convertedTrackId.String())
	// 3. If exists, pull the artists and song features
	artistChan := make(chan *spotify.FullArtist)
	audioFeaturesChan := make(chan *spotify.AudioFeatures)
	errorChan2 := make(chan error)

	go func(artistChan chan<- *spotify.FullArtist) {
		artistIds := make([]spotify.ID, len(track.Artists))
		for i, artist := range track.Artists {
			artistIds[i] = artist.ID
		}

		err := input.SpArtistService.Repo.GetArtists(artistIds, artistChan, input.Ctx)
		if err != nil {
			errorChan <- err
		}
	}(artistChan)

	go func(audioFeaturesChan chan<- *spotify.AudioFeatures) {
		err := input.SpTrackService.Repo.GetTrackAudioFeatures(input.Ctx, *convertedTrackId, audioFeaturesChan)
		if err != nil {
			errorChan2 <- err
		}
	}(audioFeaturesChan)

	var artists []*spotify.FullArtist
	var audioFeature *spotify.AudioFeatures
	artistsDone, audioFeatureDone := false, false
	for !artistsDone || !audioFeatureDone {
		select {
		case err := <-errorChan:
			close(artistChan)
			log.WithFields(input.Meta).Errorf("Error getting artists: %v", err)
			return nil, err
		case err := <-errorChan2:
			close(audioFeaturesChan)
			log.WithFields(input.Meta).Errorf("Error getting audio features: %v", err)
			return nil, err
		case artist, ok := <-artistChan:
			if artist != nil {
				artists = append(artists, artist)
			}

			if !ok {
				artistsDone = true
			}
		case audioFeature = <-audioFeaturesChan:
			if audioFeature == nil {
				log.WithFields(input.Meta).Debugf("No audio features found for track %v", convertedTrackId.String())
				return nil, domain.ErrNoTrackAudioFeatures
			}

			audioFeatureDone = true
		}
	}

	log.WithFields(input.Meta).Debugf("Finished getting artists and audio features for track %v", convertedTrackId.String())
	// 2.3 Check each of the genres for each artist and ensure that none of them are blacklisted
	for _, artist := range artists {
		for _, genre := range artist.Genres {
			genreBlacklisted, err := input.BlacklistService.Repo.CheckBlacklistItem(input.Ctx, domain.Genre, genre)
			if err != nil {
				return nil, err
			}

			if genreBlacklisted {
				log.WithFields(input.Meta).Debugf("Genre %v for artist %v is blacklisted", genre, artist.Name)
				return nil, &domain.ErrGenreBlacklisted{GenreName: genre, ArtistName: artist.Name}
			}
		}
	}

	log.WithFields(input.Meta).Debugf("Adding track %v to playlist", convertedTrackId.String())
	// 4. Add to playlist
	err = input.SpPlaylistService.Repo.AddTracksToPlaylist(input.Ctx, []spotify.ID{track.ID})
	if err != nil {
		log.WithFields(input.Meta).Errorf("Error adding track to playlist: %v", err)
		return nil, domain.ErrCouldNotAddToPlaylist
	}

	log.WithFields(input.Meta).Debugf("Adding track %v to database", convertedTrackId.String())
	// 5. Add to databases
	err = input.TrackService.Repo.AddTrackToDatabase(input.UserFields, track, artists, audioFeature)
	if err != nil {
		// Compensation steps in case of failure, since the track was already added to the playlist
		log.WithFields(input.Meta).Errorf("Error adding track to database: %v", err)

		log.WithFields(input.Meta).Debugf("Attempting to rollback adding track %v to playlist", convertedTrackId.String())

		err2 := input.SpPlaylistService.Repo.RemoveTracksFromPlaylist(input.Ctx, []spotify.ID{track.ID})
		if err2 != nil {
			log.WithFields(input.Meta).Errorf("Error removing track from playlist during rollback: %v", err2)
			return nil, domain.ErrCouldNotRemoveFromPlaylist
		}

		return nil, domain.ErrCouldNotAddToDatabase
	}

	return track, nil
}

func handleMaxDuration(preferenceService *domain.PreferenceService, track *spotify.FullTrack, meta log.Fields, convertedTrackId *spotify.ID, ctx context.Context) error {
	// 2.5 Check that the duration of the song does not exceed the maximum
	maxDurationPreference, err := preferenceService.Repo.Get(ctx, domain.MaxDurationKey)
	if err != nil {
		return err
	}
	var maxDuration int
	if v, ok := maxDurationPreference.Value.(int32); ok {
		maxDuration = int(v)
	} else {
		log.Warn("Max duration preference is not an int32, using default value")
		maxDuration = domain.MaxDurationKey.DefaultValue().(int)
	}

	if int(track.Duration) > maxDuration {
		log.WithFields(meta).Debugf("Track %v exceeds maximum duration", convertedTrackId.String())
		return domain.ErrTrackTooLong
	}

	return nil
}

func handleTrackOrArtistBlacklisted(blacklistService *domain.BlacklistService, track *spotify.FullTrack, ctx context.Context, meta log.Fields) error {
	// 2.1 Check that the track is not blacklisted
	trackBlacklisted, err := blacklistService.Repo.CheckBlacklistItem(ctx, domain.Track, track.ID.String())
	if err != nil {
		return err
	}

	if trackBlacklisted {
		log.WithFields(meta).Debugf("Track %v is blacklisted", track.Name)
		artistNames := make([]string, len(track.Artists))
		for i, artist := range track.Artists {
			artistNames[i] = artist.Name
		}
		return &domain.ErrTrackBlacklisted{TrackName: track.Name, ArtistNames: artistNames}
	}

	// 2.2 Check each of the artists and ensure that none of them are blacklisted
	for _, artist := range track.Artists {
		artistBlacklisted, err := blacklistService.Repo.CheckBlacklistItem(ctx, domain.Artist, artist.ID.String())
		if err != nil {
			return err
		}

		if artistBlacklisted {
			log.WithFields(meta).Debugf("Artist %v is blacklisted", artist.ID.String())
			return &domain.ErrArtistBlacklisted{ArtistName: artist.Name}
		}
	}

	return nil
}
