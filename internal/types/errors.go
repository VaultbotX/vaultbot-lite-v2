package types

import "errors"

var (
	ErrInvalidTrackId             = errors.New("spotify: invalid track id")
	ErrTrackAlreadyInPlaylist     = errors.New("spotify: track already exists in playlist")
	ErrNoTrackExists              = errors.New("spotify: no track exists")
	ErrNoTrackAudioFeatures       = errors.New("spotify: no track audio features")
	ErrTrackTooLong               = errors.New("spotify: track is too long")
	ErrCouldNotAddToPlaylist      = errors.New("spotify: could not add track to playlist")
	ErrCouldNotAddToDatabase      = errors.New("spotify: could not add track to database")
	ErrCouldNotRemoveFromPlaylist = errors.New("spotify: could not remove track from playlist")
	ErrUnsupportedOSForBrowser    = errors.New("unsupported OS for browser")
	ErrUnauthorized               = errors.New("spotify: unauthorized")
	ErrBlacklistItemAlreadyExists = errors.New("spotify: blacklist item already exists")
)

// ErrTrackBlacklisted TODO: handle these errors
type ErrTrackBlacklisted struct {
	TrackName   string
	ArtistNames []string
}

func (e ErrTrackBlacklisted) Error() string {
	return "spotify: track is blacklisted"
}

type ErrArtistBlacklisted struct {
	ArtistName string
}

func (e ErrArtistBlacklisted) Error() string {
	return "spotify: artist is blacklisted"
}

type ErrGenreBlacklisted struct {
	GenreName  string
	ArtistName string
}

func (e ErrGenreBlacklisted) Error() string {
	return "spotify: genre is blacklisted"
}
