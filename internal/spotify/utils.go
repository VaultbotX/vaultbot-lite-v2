package spotify

import (
	"github.com/zmb3/spotify/v2"
	"regexp"
)

var alphanumericRegex = regexp.MustCompile(`^[a-zA-Z0-9]{1,50}$`)
var spotifyUriRegex = regexp.MustCompile(`spotify:track:(\w+)`)
var spotifyUrlRegex = regexp.MustCompile(`https://open.spotify.com/track/(\w+)`)

// ParseTrackId parses a string and returns a spotify ID for a Spotify URL, URI, or track ID.
func ParseTrackId(text string) *spotify.ID {
	if alphanumericRegex.MatchString(text) {
		match := spotify.ID(text)
		return &match
	}

	if spotifyUriRegex.MatchString(text) {
		match := spotifyUriRegex.FindStringSubmatch(text)
		if len(match) > 0 {
			firstMatch := spotify.ID(match[1])
			return &firstMatch
		}
	}

	if spotifyUrlRegex.MatchString(text) {
		match := spotifyUrlRegex.FindStringSubmatch(text)
		if len(match) > 0 {
			firstMatch := spotify.ID(match[1])
			return &firstMatch
		}
	}

	return nil
}
