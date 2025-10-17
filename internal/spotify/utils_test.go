package spotify

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/vaultbotx/vaultbot-lite/internal/domain"
)

type mockClient struct {
	resp *http.Response
	err  error
}

func (m *mockClient) Do(_ *http.Request) (*http.Response, error) {
	return m.resp, m.err
}

func TestParseTrackId_ParsesAlphanumericText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"ParsesNumericText", "1234567890", "1234567890"},
		{"ParsesAlphaText", "ABCDEFG", "ABCDEFG"},
		{"ParsesAlphanumericText", "ABC123", "ABC123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := ParseSpotifyId(tt.input, domain.Track)
			if tt.expected == "" {
				if actual != nil {
					t.Errorf("Expected nil, got %s", actual)
				}
			} else {
				if actual == nil {
					t.Errorf("Expected %s, got nil", tt.expected)
				} else if actual.String() != tt.expected {
					t.Errorf("Expected %s, got %s", tt.expected, actual)
				}
			}
		})
	}
}

func TestParseTrackId_InvalidLengthString_ReturnsNil(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"EmptyString", ""},
		{"TooLongString", string(make([]byte, 51))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := ParseSpotifyId(tt.input, domain.Track)
			if actual != nil {
				t.Errorf("Expected nil, got %s", actual)
			}
		})
	}
}

func TestParseTrackId_ParsesSpotifyUri(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"FirstUri", "spotify:track:4G4patpYxsF6ovHZOX9wgR", "4G4patpYxsF6ovHZOX9wgR"},
		{"SecondUri", "spotify:track:2gQK13gXYZRq2MgvPJyHx8", "2gQK13gXYZRq2MgvPJyHx8"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := ParseSpotifyId(tt.input, domain.Track)
			if tt.expected == "" {
				if actual != nil {
					t.Errorf("Expected nil, got %s", actual)
				}
			} else {
				if actual == nil {
					t.Errorf("Expected %s, got nil", tt.expected)
				} else if actual.String() != tt.expected {
					t.Errorf("Expected %s, got %s", tt.expected, actual)
				}
			}
		})
	}
}

func TestParseTrackId_DoesNotParseOtherSpotifyEntityTypeUris(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"PodcastUri", "spotify:episode:4G4patpYxsF6ovHZOX9wgR"},
		{"AlbumUri", "spotify:album:4G4patpYxsF6ovHZOX9wgR"},
		{"ArtistUri", "spotify:artist:4G4patpYxsF6ovHZOX9wgR"},
		{"PlaylistUri", "spotify:playlist:4G4patpYxsF6ovHZOX9wgR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := ParseSpotifyId(tt.input, domain.Track)
			if actual != nil {
				t.Errorf("Expected nil, got %s", actual)
			}
		})
	}
}

func TestParseTrackId_ParsesSpotifyUrl(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"UrlWithQueryString", "https://open.spotify.com/track/2gQK13gXYZRq2MgvPJyHx8?si=67d66f6ee4e5494c", "2gQK13gXYZRq2MgvPJyHx8"},
		{"UrlWithoutQueryString", "https://open.spotify.com/track/2gQK13gXYZRq2MgvPJyHx8", "2gQK13gXYZRq2MgvPJyHx8"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := ParseSpotifyId(tt.input, domain.Track)
			if tt.expected == "" {
				if actual != nil {
					t.Errorf("Expected nil, got %s", actual)
				}
			} else {
				if actual == nil {
					t.Errorf("Expected %s, got nil", tt.expected)
				} else if actual.String() != tt.expected {
					t.Errorf("Expected %s, got %s", tt.expected, actual)
				}
			}
		})
	}
}

func TestParseSpotifyId_ShortLink_FromHTML(t *testing.T) {
	orig := httpClient
	defer func() { httpClient = orig }()

	html := `<!doctype html><html><head><title>Test</title></head><body>
	<a href="https://open.spotify.com/track/06KfJvxTq7GFzGIr0tRwPE?si=foobar">Listen on Spotify</a>
	</body></html>`

	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(html)),
		Request:    &http.Request{URL: &url.URL{Scheme: "https", Host: "spotify.app.link", Path: "/"}},
	}

	httpClient = &mockClient{resp: resp}

	id := ParseSpotifyId("https://spotify.link/short", domain.Track)
	if id == nil {
		t.Fatal("expected non-nil id")
	}
	if string(*id) != "06KfJvxTq7GFzGIr0tRwPE" {
		t.Fatalf("unexpected id: %s", string(*id))
	}
}

func TestParseSpotifyId_ShortLink_FromRespURL(t *testing.T) {
	orig := httpClient
	defer func() { httpClient = orig }()

	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader("<html></html>")),
		Request:    &http.Request{URL: &url.URL{Scheme: "https", Host: "open.spotify.com", Path: "/track/XYZ12345"}},
	}

	httpClient = &mockClient{resp: resp}

	id := ParseSpotifyId("https://spotify.link/short2", domain.Track)
	if id == nil {
		t.Fatal("expected non-nil id")
	}
	if string(*id) != "XYZ12345" {
		t.Fatalf("unexpected id: %s", string(*id))
	}
}
