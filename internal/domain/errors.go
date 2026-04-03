package domain

import "errors"

var (
	ErrInvalidSpotifyId           = errors.New("spotify: invalid spotify entity id")
	ErrNoTrackExists              = errors.New("spotify: no track exists")
	ErrCouldNotAddToPlaylist      = errors.New("spotify: could not add track to playlist")
	ErrCouldNotAddToDatabase      = errors.New("spotify: could not add track to database")
	ErrCouldNotRemoveFromPlaylist = errors.New("spotify: could not remove track from playlist")
	ErrUnauthorized               = errors.New("spotify: unauthorized")
)
