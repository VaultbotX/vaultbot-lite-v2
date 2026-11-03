package musicbrainz

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/time/rate"
)

type Client struct {
	httpClient *http.Client
	limiter    *rate.Limiter
	userAgent  string
	maxRetries int
	baseURL    string
}

type ClientOption func(*Client)

func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

func WithRateLimiter(limiter *rate.Limiter) ClientOption {
	return func(c *Client) {
		c.limiter = limiter
	}
}

func WithMaxRetries(retries int) ClientOption {
	return func(c *Client) {
		c.maxRetries = retries
	}
}

func WithUserAgent(userAgent string) ClientOption {
	return func(c *Client) {
		c.userAgent = userAgent
	}
}

func NewClient(opts ...ClientOption) *Client {
	c := &Client{
		httpClient: &http.Client{Timeout: 15 * time.Second},
		limiter:    rate.NewLimiter(rate.Every(1050*time.Millisecond), 1),
		userAgent:  "VaultbotGenreService/1.0.0 ( trey@tbrittain.com / https://github.com/VaultbotX )",
		maxRetries: 3,
		baseURL:    "https://musicbrainz.org/ws/2",
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

type urlLookupResponse struct {
	Relations []struct {
		Type   string `json:"type"`
		Artist struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"artist"`
	} `json:"relations"`
}

type artistResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Genres []struct {
		Name string `json:"name"`
	} `json:"genres"`
	Tags []struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	} `json:"tags"`
}

type searchResponse struct {
	Artists []struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Genres []struct {
			Name string `json:"name"`
		} `json:"genres"`
		Tags []struct {
			Name  string `json:"name"`
			Count int    `json:"count"`
		} `json:"tags"`
	} `json:"artists"`
}

func (c *Client) executeRequest(ctx context.Context, reqURL string, target any) error {
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if err := c.limiter.Wait(ctx); err != nil {
			return err
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
		if err != nil {
			return err
		}

		req.Header.Set("User-Agent", c.userAgent)
		req.Header.Set("Accept", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			if attempt < c.maxRetries {
				time.Sleep(time.Duration(attempt+1) * 2 * time.Second)
				continue
			}
			return err
		}

		if resp.StatusCode == http.StatusNotFound {
			_ = resp.Body.Close()
			return nil
		}

		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusServiceUnavailable {
			_ = resp.Body.Close()
			if attempt < c.maxRetries {
				time.Sleep(time.Duration(attempt+1) * 2 * time.Second)
				continue
			}
			return fmt.Errorf("musicbrainz rate limit or service unavailable (HTTP %d)", resp.StatusCode)
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			_ = resp.Body.Close()
			return fmt.Errorf("musicbrainz request failed with status: %d", resp.StatusCode)
		}

		err = json.NewDecoder(resp.Body).Decode(target)
		_ = resp.Body.Close()
		return err
	}

	return nil
}

func (c *Client) GetGenresBySpotifyID(ctx context.Context, spotifyID string) ([]string, error) {
	spotifyURL := fmt.Sprintf("https://open.spotify.com/artist/%s", spotifyID)
	reqURL := fmt.Sprintf("%s/url?resource=%s&inc=artist-rels&fmt=json", c.baseURL, url.QueryEscape(spotifyURL))

	var lookup urlLookupResponse
	if err := c.executeRequest(ctx, reqURL, &lookup); err != nil {
		return nil, err
	}

	var artistID string
	for _, rel := range lookup.Relations {
		if rel.Artist.ID != "" && rel.Type == "free streaming" {
			artistID = rel.Artist.ID
			break
		}
	}

	if artistID == "" && len(lookup.Relations) > 0 {
		artistID = lookup.Relations[0].Artist.ID
	}

	if artistID == "" {
		return nil, nil
	}

	artistURL := fmt.Sprintf("%s/artist/%s?inc=genres+tags&fmt=json", c.baseURL, artistID)
	var artist artistResponse
	if err := c.executeRequest(ctx, artistURL, &artist); err != nil {
		return nil, err
	}

	var genres []string
	for _, g := range artist.Genres {
		if g.Name != "" {
			genres = append(genres, g.Name)
		}
	}

	if len(genres) == 0 {
		for _, t := range artist.Tags {
			if t.Name != "" {
				genres = append(genres, t.Name)
			}
		}
	}

	return genres, nil
}

func (c *Client) GetGenresByName(ctx context.Context, name string) ([]string, error) {
	reqURL := fmt.Sprintf("%s/artist/?query=artist:%s&fmt=json", c.baseURL, url.QueryEscape(name))

	var search searchResponse
	if err := c.executeRequest(ctx, reqURL, &search); err != nil {
		return nil, err
	}

	if len(search.Artists) == 0 {
		return nil, nil
	}

	targetName := strings.ToLower(strings.TrimSpace(name))
	var matchedArtist *struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Genres []struct {
			Name string `json:"name"`
		} `json:"genres"`
		Tags []struct {
			Name  string `json:"name"`
			Count int    `json:"count"`
		} `json:"tags"`
	}

	for i := range search.Artists {
		if strings.ToLower(strings.TrimSpace(search.Artists[i].Name)) == targetName {
			matchedArtist = &search.Artists[i]
			break
		}
	}

	if matchedArtist == nil {
		matchedArtist = &search.Artists[0]
	}

	var genres []string
	for _, g := range matchedArtist.Genres {
		if g.Name != "" {
			genres = append(genres, g.Name)
		}
	}

	if len(genres) == 0 {
		for _, t := range matchedArtist.Tags {
			if t.Name != "" {
				genres = append(genres, t.Name)
			}
		}
	}

	return genres, nil
}
