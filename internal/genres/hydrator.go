package genres

import (
	"context"

	log "github.com/sirupsen/logrus"
)

type MusicBrainzSource interface {
	GetGenresBySpotifyID(ctx context.Context, spotifyID string) ([]string, error)
	GetGenresByName(ctx context.Context, name string) ([]string, error)
}

type WikidataSource interface {
	GetGenresBySpotifyID(ctx context.Context, spotifyID string) ([]string, error)
	GetGenresByName(ctx context.Context, name string) ([]string, error)
}

type Hydrator struct {
	musicBrainz MusicBrainzSource
	wikidata    WikidataSource
}

func NewHydrator(mb MusicBrainzSource, wd WikidataSource) *Hydrator {
	return &Hydrator{
		musicBrainz: mb,
		wikidata:    wd,
	}
}

type GenreSource interface {
	GetGenresBySpotifyID(ctx context.Context, spotifyID string) ([]string, error)
	GetGenresByName(ctx context.Context, name string) ([]string, error)
}

func (h *Hydrator) querySource(ctx context.Context, source GenreSource, spotifyID string, name string) []string {
	if spotifyID != "" {
		genres, err := source.GetGenresBySpotifyID(ctx, spotifyID)
		if err != nil {
			log.Warnf("Failed to query genres by spotify ID %s: %v", spotifyID, err)
		} else if len(genres) > 0 {
			normalized := NormalizeGenres(genres)
			if len(normalized) > 0 {
				return normalized
			}
		}
	}

	if name != "" {
		genres, err := source.GetGenresByName(ctx, name)
		if err != nil {
			log.Warnf("Failed to query genres by name %s: %v", name, err)
		} else if len(genres) > 0 {
			normalized := NormalizeGenres(genres)
			if len(normalized) > 0 {
				return normalized
			}
		}
	}

	return nil
}

// HydrateArtistGenres resolves artist genres by trying MusicBrainz first, falling back to Wikidata.
func (h *Hydrator) HydrateArtistGenres(ctx context.Context, spotifyID string, artistName string) (genres []string, sourceName string) {
	if h.musicBrainz != nil {
		mbGenres := h.querySource(ctx, h.musicBrainz, spotifyID, artistName)
		if len(mbGenres) > 0 {
			return mbGenres, "musicbrainz"
		}
	}

	if h.wikidata != nil {
		wdGenres := h.querySource(ctx, h.wikidata, spotifyID, artistName)
		if len(wdGenres) > 0 {
			return wdGenres, "wikidata"
		}
	}

	return nil, ""
}
