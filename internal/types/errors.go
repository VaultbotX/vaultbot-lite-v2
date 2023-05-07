package types

import "errors"

var (
	ErrInvalidTrackId             = errors.New("spotify: invalid track id")
	ErrTrackAlreadyInPlaylist     = errors.New("spotify: track already exists in playlist")
	ErrNoTrackExists              = errors.New("spotify: no track exists")
	ErrTrackTooLong               = errors.New("spotify: track is too long")
	ErrCouldNotAddToPlaylist      = errors.New("spotify: could not add track to playlist")
	ErrCouldNotAddToDatabase      = errors.New("spotify: could not add track to database")
	ErrCouldNotRemoveFromPlaylist = errors.New("spotify: could not remove track from playlist")
	ErrUnsupportedOSForBrowser    = errors.New("unsupported OS for browser")
)
