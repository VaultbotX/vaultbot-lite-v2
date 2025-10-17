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
