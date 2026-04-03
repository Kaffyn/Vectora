package index

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Client struct {
	BaseURL string
	HTTP    *http.Client
}

func NewClient() *Client {
	return &Client{
		BaseURL: "https://index.vectora.org/v1", // Canonical O.S.S Registry
		HTTP:    &http.Client{Timeout: 60 * time.Second}, // Tolerate slow mirrors
	}
}

// DownloadArchive fetches heavy GGUFs or Snapshots in Chunks to avoid RAM explosion.
// Uses an 'onProgress' callback that the Motherboard (internal/ipc/server.go) attaches to the Broadcast.
func (c *Client) DownloadArchive(ctx context.Context, id string, destPath string, onProgress func(percent float64)) error {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/datasets/%s/download", c.BaseURL, id), nil)
	if err != nil {
		return err
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("index_client_err: Server responded with code %d", resp.StatusCode)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Intelligent Stream Reader
	totalBytes := resp.ContentLength
	var downloaded int64

	buf := make([]byte, 32*1024) // Chunks limited to 32KB
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		n, err := resp.Body.Read(buf)
		if n > 0 {
			out.Write(buf[:n])
			downloaded += int64(n)
			
			if totalBytes > 0 && onProgress != nil {
				pct := (float64(downloaded) / float64(totalBytes)) * 100.0
				onProgress(pct)
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	return nil
}
