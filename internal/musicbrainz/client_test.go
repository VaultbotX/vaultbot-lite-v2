package musicbrainz

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/time/rate"
)

func TestMusicBrainzClient_GetGenresBySpotifyID(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Errorf("expected User-Agent header to be set")
		}

		if strings.Contains(r.URL.Path, "/url") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"relations": [
					{
						"type": "free streaming",
						"artist": {
							"id": "mbid-123",
							"name": "Chappell Roan"
						}
					}
				]
			}`))
			return
		}

		if strings.Contains(r.URL.Path, "/artist/mbid-123") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"id": "mbid-123",
				"name": "Chappell Roan",
				"genres": [{"name": "synth-pop"}, {"name": "dance-pop"}],
				"tags": []
			}`))
			return
		}

		http.NotFound(w, r)
	}))
	defer ts.Close()

	client := NewClient(
		WithBaseURL(ts.URL),
		WithRateLimiter(rate.NewLimiter(rate.Inf, 1)),
	)

	genres, err := client.GetGenresBySpotifyID(context.Background(), "7GlBOeep6PqTfFi59PTUUN")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(genres) != 2 || genres[0] != "synth-pop" || genres[1] != "dance-pop" {
		t.Errorf("unexpected genres: %v", genres)
	}
}

func TestMusicBrainzClient_GetGenresByName(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"artists": [
				{
					"id": "mbid-cannons",
					"name": "Cannons",
					"genres": [{"name": "indie pop"}, {"name": "electropop"}]
				}
			]
		}`))
	}))
	defer ts.Close()

	client := NewClient(
		WithBaseURL(ts.URL),
		WithRateLimiter(rate.NewLimiter(rate.Inf, 1)),
	)

	genres, err := client.GetGenresByName(context.Background(), "Cannons")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(genres) != 2 || genres[0] != "indie pop" || genres[1] != "electropop" {
		t.Errorf("unexpected genres: %v", genres)
	}
}
