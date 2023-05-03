package types

import "errors"

var ErrInvalidTrackId = errors.New("spotify: invalid track id")

var ErrTrackAlreadyInPlaylist = errors.New("spotify: track already exists in playlist")

var ErrNoTrackExists = errors.New("spotify: no track exists")

var ErrCouldNotAddToPlaylist = errors.New("spotify: could not add track to playlist")

var ErrCouldNotAddToDatabase = errors.New("spotify: could not add track to database")

var ErrCouldNotRemoveFromPlaylist = errors.New("spotify: could not remove track from playlist")