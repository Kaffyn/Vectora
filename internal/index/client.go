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
		BaseURL: "https://index.vectora.org/v1", // Registry canônico O.S.S
		HTTP:    &http.Client{Timeout: 60 * time.Second}, // Tolera mirrors lentos
	}
}

// DownloadArchive traz GGUFs ou Snapshots pesados em Chunks pra não explodir a RAM.
// Utiliza um callback 'onProgress' que a Placa Mãe (internal/ipc/server.go) atrela ao Broadcast.
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
		return fmt.Errorf("index_client_err: Servidor respondeu com código %d", resp.StatusCode)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Leitor de Stream Inteligente
	totalBytes := resp.ContentLength
	var downloaded int64

	buf := make([]byte, 32*1024) // Chunks limitados em 32KB
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
