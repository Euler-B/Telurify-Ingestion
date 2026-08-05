package usgs

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const FeedURL = "https://earthquake.usgs.gov/earthquakes/feed/v1.0/summary/all_month.geojson"

const maxFeedBodySize = 10 << 20 // 10 MiB

func Fetch() (*FeatureCollection, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return fmt.Errorf("redirects are not allowed")
		},
	}

	resp, err := client.Get(FeedURL)
	if err != nil {
		return nil, fmt.Errorf("fetching USGS feed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("USGS feed returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxFeedBodySize+1))
	if err != nil {
		return nil, fmt.Errorf("reading USGS response body: %w", err)
	}
	if len(body) > maxFeedBodySize {
		return nil, fmt.Errorf("USGS response body exceeds %d bytes", maxFeedBodySize)
	}

	var fc FeatureCollection
	if err := json.Unmarshal(body, &fc); err != nil {
		return nil, fmt.Errorf("parsing USGS GeoJSON: %w", err)
	}

	return &fc, nil
}
