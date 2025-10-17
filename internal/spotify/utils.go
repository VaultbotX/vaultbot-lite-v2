package spotify

import (
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/vaultbotx/vaultbot-lite/internal/domain"
	"github.com/zmb3/spotify/v2"
)

var alphanumericRegex = regexp.MustCompile(`^[a-zA-Z0-9]{1,50}$`)
var spotifyTrackUriRegex = regexp.MustCompile(`spotify:track:(\w+)`)
var spotifyArtistUriRegex = regexp.MustCompile(`spotify:artist:(\w+)`)
var spotifyTrackUrlRegex = regexp.MustCompile(`^https://open\.spotify\.com/track/(\w+)(\?.*)?$`)
var spotifyArtistUrlRegex = regexp.MustCompile(`^https://open\.spotify\.com/artist/(\w+)(\?.*)?$`)
var spotifyShortLinkRegex = regexp.MustCompile(`^https?://spotify\.link/[A-Za-z0-9_-]+$`)

// hrefRegex matches href attributes (single-quoted, double-quoted, or unquoted) and is reused by resolveSpotifyLink.
var hrefRegex = regexp.MustCompile(`href\s*=\s*(?:'([^']*)'|"([^"]*)"|([^>\s]+))`)

// httpClientDo models the subset of http.Client we need (Do method). Tests can replace httpClient.
type httpClientDo interface {
	Do(req *http.Request) (*http.Response, error)
}

// isSpotifyHost returns true for spotify.com and its subdomains.
func isSpotifyHost(host string) bool {
	if host == "spotify.com" {
		return true
	}
	return strings.HasSuffix(host, ".spotify.com")
}

// httpClient is package-level client used for network calls; replace in tests to avoid real HTTP.
var httpClient httpClientDo = &http.Client{
	Timeout: 15 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		// Only allow redirects where the target hostname is spotify.com or a subdomain of spotify.com.
		// This prevents following redirects to attacker-controlled domains.
		h := req.URL.Hostname()
		if isSpotifyHost(h) {
			return nil
		}
		// otherwise, do not follow the redirect; return ErrUseLastResponse to return the previous response
		return http.ErrUseLastResponse
	},
}

// resolveSpotifyLink fetches the given spotify.link (or other intermediary) and attempts to find
// an open.spotify.com link within the returned HTML. It returns the first found open.spotify.com URL
// or an empty string if none found or on error.
func resolveSpotifyLink(startURL string) string {
	if startURL == "" {
		return ""
	}

	req, err := http.NewRequest("GET", startURL, nil)
	if err != nil {
		return ""
	}
	// set a conservative User-Agent
	req.Header.Set("User-Agent", "vaultbot-spotify-resolver/1.0")

	resp, err := httpClient.Do(req)
	if err != nil {
		return ""
	}
	defer func() { _ = resp.Body.Close() }()

	maxBodySize := int64(1 << 20) // 1 MiB
	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, maxBodySize))
	if err != nil {
		return ""
	}
	body := string(bodyBytes)

	// quick scan for hrefs that point to open.spotify.com
	// crude but practical regex-like search: look for open.spotify.com in hrefs
	// We'll search for href= patterns and resolve them relative to startURL using simple string resolution.
	matches := hrefRegex.FindAllStringSubmatch(body, -1)
	for _, m := range matches {
		raw := ""
		if len(m) >= 2 && m[1] != "" {
			raw = m[1]
		} else if len(m) >= 3 && m[2] != "" {
			raw = m[2]
		} else if len(m) >= 4 {
			raw = m[3]
		}
		if raw == "" {
			continue
		}
		// normalize and strip obvious trailing punctuation
		raw = strings.TrimSpace(raw)
		raw = strings.TrimRight(raw, `"'`)
		// Parse the href and resolve relative URLs against the response request URL if necessary.
		u, err := url.Parse(raw)
		if err != nil {
			continue
		}
		if !u.IsAbs() && resp.Request != nil && resp.Request.URL != nil {
			u = resp.Request.URL.ResolveReference(u)
		}
		// validate scheme and host precisely
		if (u.Scheme == "http" || u.Scheme == "https") && u.Hostname() == "open.spotify.com" {
			return u.String()
		}
	}

	// as a fallback, if the final request url itself is on open.spotify.com, return it
	if resp.Request != nil && resp.Request.URL != nil && resp.Request.URL.Hostname() == "open.spotify.com" {
		return resp.Request.URL.String()
	}

	return ""
}

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

	// If the text is a spotify.link short URL, attempt to resolve it to an open.spotify.com URL
	if spotifyShortLinkRegex.MatchString(text) {
		resolved := resolveSpotifyLink(text)
		if resolved != "" {
			text = resolved
		}
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
