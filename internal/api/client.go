package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"strings"
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

func (c *Client) GetSnapshot(ctx context.Context, origin string, snapshot string) (*Snapshot, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/api/v1/snapshots/%s/%s", c.Endpoint, origin, snapshot), nil)
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

	var result Snapshot
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) CopyBlob(ctx context.Context, w http.ResponseWriter, digest string) error {
	algorithm, digest, ok := strings.Cut(digest, ":")
	if !ok {
		return fmt.Errorf("invalid digest")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/api/v1/blobs/%s/%s", c.Endpoint, algorithm, digest), nil)
	if err != nil {
		return err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	maps.Copy(w.Header(), res.Header)
	w.WriteHeader(res.StatusCode)

	_, err = io.Copy(w, res.Body)
	return err
}
