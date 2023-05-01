package types

import "errors"

var ErrNoTrackExists = errors.New("spotify: no track exists")

var ErrCouldNotAddToPlaylist = errors.New("spotify: could not add track to playlist")

var ErrCouldNotAddToDatabase = errors.New("spotify: could not add track to database")

var ErrCouldNotRemoveFromPlaylist = errors.New("spotify: could not remove track from playlist")
