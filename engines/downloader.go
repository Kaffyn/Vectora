package engines

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// Downloader gerencia downloads de builds com suporte a resume.
type Downloader struct {
	client    *http.Client
	timeout   time.Duration
	maxRetries int
}

// NewDownloader cria um novo downloader com timeouts padrão.
func NewDownloader() *Downloader {
	return &Downloader{
		client: &http.Client{
			Timeout: 30 * time.Minute,
		},
		timeout:    30 * time.Minute,
		maxRetries: 3,
	}
}

// Download baixa um arquivo com suporte a resume e callbacks de progresso.
//
// Se o arquivo já existe parcialmente (.partial), tenta retomar o download.
// Ao completar, renomeia .partial para o nome final.
func (d *Downloader) Download(
	ctx context.Context,
	url string,
	destPath string,
	onProgress func(*DownloadProgress) error,
) error {
	// Criar diretório de destino se necessário
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	partialPath := destPath + ".partial"

	// Verificar se há download parcial anterior
	var startByte int64 = 0
	fileInfo, err := os.Stat(partialPath)
	if err == nil {
		startByte = fileInfo.Size()
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to stat partial file: %w", err)
	}

	// HEAD request para obter tamanho total
	totalSize, supportsRange, err := d.getFileSize(ctx, url)
	if err != nil {
		return fmt.Errorf("failed to get file size: %w", err)
	}

	// Se não suporta range e há arquivo parcial, limpar e começar do zero
	if !supportsRange && startByte > 0 {
		os.Remove(partialPath)
		startByte = 0
	}

	// Tentar download com retry
	var lastErr error
	for attempt := 0; attempt < d.maxRetries; attempt++ {
		lastErr = d.downloadWithResume(ctx, url, partialPath, startByte, totalSize, onProgress)
		if lastErr == nil {
			break
		}

		// Incrementar byte de início para o próximo retry
		fileInfo, err := os.Stat(partialPath)
		if err == nil {
			startByte = fileInfo.Size()
		}

		// Exponential backoff
		backoff := time.Duration((attempt + 1) * 2) * time.Second
		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	if lastErr != nil {
		return lastErr
	}

	// Renomear .partial para final
	if err := os.Rename(partialPath, destPath); err != nil {
		return fmt.Errorf("failed to finalize download: %w", err)
	}

	return nil
}

// downloadWithResume faz o download propriamente dito.
func (d *Downloader) downloadWithResume(
	ctx context.Context,
	url string,
	filePath string,
	startByte int64,
	totalSize int64,
	onProgress func(*DownloadProgress) error,
) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Range header se há arquivo parcial
	if startByte > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", startByte))
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("download request failed: %w", err)
	}
	defer resp.Body.Close()

	// Verificar status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Tamanho da resposta é fornecido pela resposta anterior (totalSize)

	// Abrir arquivo para escrita/append
	flags := os.O_CREATE | os.O_WRONLY
	if startByte > 0 {
		flags |= os.O_APPEND
	}

	f, err := os.OpenFile(filePath, flags, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file for writing: %w", err)
	}
	defer f.Close()

	// Download com progresso
	bytesDownloaded := startByte
	lastProgressTime := time.Now()
	lastProgressBytes := startByte

	// Buffer para leitura
	buf := make([]byte, 32*1024) // 32KB chunks

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, err := f.Write(buf[:n]); err != nil {
				return fmt.Errorf("failed to write to file: %w", err)
			}
			bytesDownloaded += int64(n)

			// Relatório de progresso a cada 100ms
			now := time.Now()
			if now.Sub(lastProgressTime) > 100*time.Millisecond {
				speed := float64(bytesDownloaded-lastProgressBytes) / now.Sub(lastProgressTime).Seconds()
				progress := &DownloadProgress{
					Current: bytesDownloaded,
					Total:   totalSize,
					Speed:   speed,
				}
				if onProgress != nil {
					if err := onProgress(progress); err != nil {
						return fmt.Errorf("progress callback error: %w", err)
					}
				}
				lastProgressTime = now
				lastProgressBytes = bytesDownloaded
			}
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("read error: %w", err)
		}
	}

	return nil
}

// getFileSize faz um HEAD request para obter tamanho total e verificar suporte a Range.
func (d *Downloader) getFileSize(ctx context.Context, url string) (int64, bool, error) {
	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return 0, false, fmt.Errorf("failed to create HEAD request: %w", err)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return 0, false, fmt.Errorf("HEAD request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	contentLength, _ := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
	supportsRange := resp.Header.Get("Accept-Ranges") == "bytes"

	return contentLength, supportsRange, nil
}
