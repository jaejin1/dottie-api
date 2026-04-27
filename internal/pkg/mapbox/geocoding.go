package mapbox

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type GeocodingResult struct {
	PlaceName string
	Category  string
}

type GeocodingClient interface {
	ReverseGeocode(ctx context.Context, lat, lng float64) (*GeocodingResult, error)
}

type HTTPClient struct {
	accessToken string
	httpClient  *http.Client
}

func NewClient(accessToken string) *HTTPClient {
	return &HTTPClient{
		accessToken: accessToken,
		httpClient:  &http.Client{},
	}
}

func (c *HTTPClient) ReverseGeocode(ctx context.Context, lat, lng float64) (*GeocodingResult, error) {
	url := fmt.Sprintf(
		"https://api.mapbox.com/geocoding/v5/mapbox.places/%f,%f.json?access_token=%s&language=ko&types=poi,address",
		lng, lat, c.accessToken,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Features []struct {
			PlaceName  string `json:"place_name"`
			Properties struct {
				Category string `json:"category"`
			} `json:"properties"`
		} `json:"features"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Features) == 0 {
		return nil, nil
	}
	return &GeocodingResult{
		PlaceName: result.Features[0].PlaceName,
		Category:  result.Features[0].Properties.Category,
	}, nil
}
