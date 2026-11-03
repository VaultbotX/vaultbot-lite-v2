package genres

import (
	"context"
	"reflect"
	"testing"
)

type mockSource struct {
	bySpotifyFn func(ctx context.Context, spotifyID string) ([]string, error)
	byNameFn    func(ctx context.Context, name string) ([]string, error)
}

func (m *mockSource) GetGenresBySpotifyID(ctx context.Context, spotifyID string) ([]string, error) {
	if m.bySpotifyFn != nil {
		return m.bySpotifyFn(ctx, spotifyID)
	}
	return nil, nil
}

func (m *mockSource) GetGenresByName(ctx context.Context, name string) ([]string, error) {
	if m.byNameFn != nil {
		return m.byNameFn(ctx, name)
	}
	return nil, nil
}

func TestHydrator_MusicBrainzPrimary(t *testing.T) {
	mb := &mockSource{
		bySpotifyFn: func(ctx context.Context, spotifyID string) ([]string, error) {
			return []string{"Synth-Pop", "Pop"}, nil
		},
	}
	wd := &mockSource{
		bySpotifyFn: func(ctx context.Context, spotifyID string) ([]string, error) {
			t.Errorf("wikidata should not be called if musicbrainz succeeds")
			return nil, nil
		},
	}

	h := NewHydrator(mb, wd)
	genres, source := h.HydrateArtistGenres(context.Background(), "spotify-1", "Chappell Roan")

	expected := []string{"synth-pop", "pop"}
	if !reflect.DeepEqual(genres, expected) || source != "musicbrainz" {
		t.Errorf("got (%v, %s), want (%v, musicbrainz)", genres, source, expected)
	}
}

func TestHydrator_WikidataFallback(t *testing.T) {
	mb := &mockSource{
		bySpotifyFn: func(ctx context.Context, spotifyID string) ([]string, error) {
			return nil, nil
		},
		byNameFn: func(ctx context.Context, name string) ([]string, error) {
			return nil, nil
		},
	}
	wd := &mockSource{
		bySpotifyFn: func(ctx context.Context, spotifyID string) ([]string, error) {
			return []string{"Electropop", "Indie Pop"}, nil
		},
	}

	h := NewHydrator(mb, wd)
	genres, source := h.HydrateArtistGenres(context.Background(), "spotify-2", "Cannons")

	expected := []string{"electropop", "indie pop"}
	if !reflect.DeepEqual(genres, expected) || source != "wikidata" {
		t.Errorf("got (%v, %s), want (%v, wikidata)", genres, source, expected)
	}
}

func TestHydrator_Unresolved(t *testing.T) {
	mb := &mockSource{}
	wd := &mockSource{}

	h := NewHydrator(mb, wd)
	genres, source := h.HydrateArtistGenres(context.Background(), "spotify-3", "Unknown")

	if len(genres) != 0 || source != "" {
		t.Errorf("expected empty, got (%v, %s)", genres, source)
	}
}
