package spotify

import (
	"github.com/vaultbotx/vaultbot-lite/internal/domain"
	"github.com/zmb3/spotify/v2"
	"regexp"
)

var alphanumericRegex = regexp.MustCompile(`^[a-zA-Z0-9]{1,50}$`)
var spotifyTrackUriRegex = regexp.MustCompile(`spotify:track:(\w+)`)
var spotifyArtistUriRegex = regexp.MustCompile(`spotify:artist:(\w+)`)
var spotifyTrackUrlRegex = regexp.MustCompile(`^https://open\.spotify\.com/track/(\w+)(\?.*)?$`)
var spotifyArtistUrlRegex = regexp.MustCompile(`^https://open\.spotify\.com/artist/(\w+)(\?.*)?$`)

// ParseSpotifyId parses a string and returns a spotify ID for a Spotify URL, URI, or track/artist ID.
// It does not handle genres, which do not have externally facing Spotify IDs.
func ParseSpotifyId(text string, entityType domain.EntityType) *spotify.ID {
	if entityType == domain.Genre {
		return nil
	}

	if alphanumericRegex.MatchString(text) {
		match := spotify.ID(text)
		return &match
	}

	switch entityType {
	case domain.Track:
		if spotifyTrackUriRegex.MatchString(text) {
			match := spotifyTrackUriRegex.FindStringSubmatch(text)
			if len(match) > 0 {
				firstMatch := spotify.ID(match[1])
				return &firstMatch
			}
		}

		if spotifyTrackUrlRegex.MatchString(text) {
			match := spotifyTrackUrlRegex.FindStringSubmatch(text)
			if len(match) > 0 {
				firstMatch := spotify.ID(match[1])
				return &firstMatch
			}
		}
		break
	case domain.Artist:
		if spotifyArtistUriRegex.MatchString(text) {
			match := spotifyArtistUriRegex.FindStringSubmatch(text)
			if len(match) > 0 {
				firstMatch := spotify.ID(match[1])
				return &firstMatch
			}
		}

		if spotifyArtistUrlRegex.MatchString(text) {
			match := spotifyArtistUrlRegex.FindStringSubmatch(text)
			if len(match) > 0 {
				firstMatch := spotify.ID(match[1])
				return &firstMatch
			}
		}
		break
	default:
		panic("this should be unreachable")
	}

	return nil
}
