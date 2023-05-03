package spotify

import (
	"github.com/zmb3/spotify/v2"
	"regexp"
)

var alphanumericRegex = regexp.MustCompile(`^[a-zA-Z0-9]{1,50}$`)

// ParseTrackId parses a string and returns a spotify ID for a Spotify URL, URI, or track ID.
func ParseTrackId(text string) *spotify.ID {
	if alphanumericRegex.MatchString(text) {
		match := spotify.ID(text)
		return &match
	}

	match := alphanumericRegex.FindAllStringSubmatch(text, -1)
	if len(match) > 0 {
		// assume the first match is the track id
		firstMatch := spotify.ID(match[0][0])
		return &firstMatch
	}

	return nil
}
