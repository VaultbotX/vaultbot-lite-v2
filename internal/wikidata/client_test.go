package wikidata

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/time/rate"
)

func TestWikidataClient_GetGenresBySpotifyID(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Errorf("expected User-Agent header")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"results": {
				"bindings": [
					{"genreLabel": {"type": "literal", "value": "synth-pop"}},
					{"genreLabel": {"type": "literal", "value": "pop music"}}
				]
			}
		}`))
	}))
	defer ts.Close()

	client := NewClient(
		WithEndpoints(ts.URL, ts.URL),
		WithRateLimiter(rate.NewLimiter(rate.Inf, 1)),
	)

	genres, err := client.GetGenresBySpotifyID(context.Background(), "7GlBOeep6PqTfFi59PTUUN")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(genres) != 2 || genres[0] != "synth-pop" || genres[1] != "pop music" {
		t.Errorf("unexpected genres: %v", genres)
	}
}

func TestWikidataClient_GetGenresByName(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "sparql") || r.URL.Query().Get("query") != "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"results": {
					"bindings": [
						{"genreLabel": {"type": "literal", "value": "indie pop"}}
					]
				}
			}`))
			return
		}

		// Search endpoint
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"search": [
				{
					"id": "Q501766",
					"label": "Cannons",
					"description": "city in Tennessee"
				},
				{
					"id": "Q100000",
					"label": "Cannons",
					"description": "American indie pop band"
				}
			]
		}`))
	}))
	defer ts.Close()

	client := NewClient(
		WithEndpoints(ts.URL, ts.URL),
		WithRateLimiter(rate.NewLimiter(rate.Inf, 1)),
	)

	genres, err := client.GetGenresByName(context.Background(), "Cannons")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(genres) != 1 || genres[0] != "indie pop" {
		t.Errorf("unexpected genres: %v", genres)
	}
}
