package wikidata

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
	httpClient     *http.Client
	limiter        *rate.Limiter
	userAgent      string
	maxRetries     int
	sparqlEndpoint string
	apiEndpoint    string
}

type ClientOption func(*Client)

func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

func WithEndpoints(sparqlEndpoint, apiEndpoint string) ClientOption {
	return func(c *Client) {
		c.sparqlEndpoint = sparqlEndpoint
		c.apiEndpoint = apiEndpoint
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
		httpClient:     &http.Client{Timeout: 15 * time.Second},
		limiter:        rate.NewLimiter(rate.Every(200*time.Millisecond), 1),
		userAgent:      "VaultbotGenreService/1.0.0 ( trey@tbrittain.com )",
		maxRetries:     3,
		sparqlEndpoint: "https://query.wikidata.org/sparql",
		apiEndpoint:    "https://www.wikidata.org/w/api.php",
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

type sparqlResponse struct {
	Results struct {
		Bindings []struct {
			GenreLabel struct {
				Type  string `json:"type"`
				Value string `json:"value"`
			} `json:"genreLabel"`
		} `json:"bindings"`
	} `json:"results"`
}

type searchResponse struct {
	Search []struct {
		ID          string `json:"id"`
		Label       string `json:"label"`
		Description string `json:"description"`
	} `json:"search"`
}

func (c *Client) executeRequest(ctx context.Context, reqURL string, acceptHeader string, target any) error {
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if err := c.limiter.Wait(ctx); err != nil {
			return err
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
		if err != nil {
			return err
		}

		req.Header.Set("User-Agent", c.userAgent)
		req.Header.Set("Accept", acceptHeader)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			if attempt < c.maxRetries {
				time.Sleep(time.Duration(attempt+1) * 1500 * time.Millisecond)
				continue
			}
			return err
		}

		if resp.StatusCode == http.StatusNotFound {
			_ = resp.Body.Close()
			return nil
		}

		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusInternalServerError || resp.StatusCode == http.StatusServiceUnavailable {
			_ = resp.Body.Close()
			if attempt < c.maxRetries {
				time.Sleep(time.Duration(attempt+1) * 1500 * time.Millisecond)
				continue
			}
			return fmt.Errorf("wikidata rate limit or service error (HTTP %d)", resp.StatusCode)
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			_ = resp.Body.Close()
			return fmt.Errorf("wikidata request failed with status: %d", resp.StatusCode)
		}

		err = json.NewDecoder(resp.Body).Decode(target)
		_ = resp.Body.Close()
		return err
	}

	return nil
}

func (c *Client) QuerySparql(ctx context.Context, sparqlQuery string) ([]string, error) {
	reqURL := fmt.Sprintf("%s?query=%s&format=json", c.sparqlEndpoint, url.QueryEscape(sparqlQuery))

	var resp sparqlResponse
	if err := c.executeRequest(ctx, reqURL, "application/sparql-results+json", &resp); err != nil {
		return nil, err
	}

	var genres []string
	for _, b := range resp.Results.Bindings {
		label := strings.TrimSpace(b.GenreLabel.Value)
		if label != "" && !strings.HasPrefix(label, "Q") && !sliceContains(genres, label) {
			genres = append(genres, label)
		}
	}

	return genres, nil
}

func (c *Client) GetGenresBySpotifyID(ctx context.Context, spotifyID string) ([]string, error) {
	query := fmt.Sprintf(`
SELECT ?genreLabel WHERE {
  ?item wdt:P1902 "%s" .
  ?item wdt:P136 ?genre .
  ?genre wdt:P31*/wdt:P279* wd:Q188451 .
  SERVICE wikibase:label { bd:serviceParam wikibase:language "en". }
}
`, spotifyID)

	return c.QuerySparql(ctx, query)
}

func (c *Client) GetGenresByName(ctx context.Context, name string) ([]string, error) {
	searchURL := fmt.Sprintf("%s?action=wbsearchentities&search=%s&language=en&type=item&format=json",
		c.apiEndpoint, url.QueryEscape(name))

	var search searchResponse
	if err := c.executeRequest(ctx, searchURL, "application/json", &search); err != nil {
		return nil, err
	}

	if len(search.Search) == 0 {
		return nil, nil
	}

	keywords := []string{"singer", "band", "musician", "musical", "group", "artist", "rapper", "dj", "composer", "producer"}
	var targetEntityID string

	for _, item := range search.Search {
		desc := strings.ToLower(item.Description)
		for _, kw := range keywords {
			if strings.Contains(desc, kw) {
				targetEntityID = item.ID
				break
			}
		}
		if targetEntityID != "" {
			break
		}
	}

	if targetEntityID == "" {
		return nil, nil
	}

	query := fmt.Sprintf(`
SELECT ?genreLabel WHERE {
  wd:%s wdt:P136 ?genre .
  ?genre wdt:P31*/wdt:P279* wd:Q188451 .
  SERVICE wikibase:label { bd:serviceParam wikibase:language "en". }
}
`, targetEntityID)

	return c.QuerySparql(ctx, query)
}

func sliceContains(slice []string, val string) bool {
	for _, item := range slice {
		if strings.EqualFold(item, val) {
			return true
		}
	}
	return false
}
