package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	Endpoint string
}

func (c *Client) GetSnapshots(ctx context.Context) (*Page[SnapshotPageEmbedded], error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.Endpoint+"/api/v1/snapshots", nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", err)
	}

	var result Page[SnapshotPageEmbedded]
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}
