package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Release represents a GitHub release with download assets.
type Release struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

// Asset represents a release asset (binary file).
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// Updater manages the update lifecycle: check, download, swap, verify, rollback.
type Updater struct {
	currentVersion string
	binaryPath     string
	backupDir      string
}

// NewUpdater creates an updater instance.
// binaryPath should be the full path to the currently running executable.
// backupDir is where old binaries are stored for rollback.
func NewUpdater(currentVersion, binaryPath, backupDir string) *Updater {
	return &Updater{
		currentVersion: currentVersion,
		binaryPath:     binaryPath,
		backupDir:      backupDir,
	}
}

// CheckForUpdates queries GitHub for the latest release.
func (u *Updater) CheckForUpdates(ctx context.Context) (*Release, error) {
	url := "https://api.github.com/repos/Kaffyn/Vectora/releases/latest"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Add User-Agent header (GitHub API requires it)
	req.Header.Set("User-Agent", "Vectora-Updater")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api returned status %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

// FindAssetForPlatform finds the correct binary asset for the current OS/arch.
func (u *Updater) FindAssetForPlatform(release *Release) *Asset {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	// Map Go's GOOS/GOARCH to likely asset names
	// Examples: vectora-windows-amd64.exe, vectora-linux-amd64, vectora-darwin-arm64
	targetPrefix := fmt.Sprintf("vectora-%s-%s", goos, goarch)
	if goos == "windows" {
		targetPrefix += ".exe"
	}

	for i := range release.Assets {
		if strings.Contains(release.Assets[i].Name, targetPrefix) {
			return &release.Assets[i]
		}
	}

	return nil
}

// DownloadAndSwap downloads the new binary and swaps it with the current one.
// Returns the path to the backup binary (for potential rollback).
func (u *Updater) DownloadAndSwap(ctx context.Context, asset *Asset) (string, error) {
	// Download new binary to temporary file
	tempFile := u.binaryPath + ".tmp"
	if err := downloadFile(ctx, asset.BrowserDownloadURL, tempFile); err != nil {
		return "", fmt.Errorf("download failed: %w", err)
	}

	// Make temporary file executable on Unix
	if runtime.GOOS != "windows" {
		_ = os.Chmod(tempFile, 0755)
	}

	// Create backup directory if needed
	_ = os.MkdirAll(u.backupDir, 0700)

	// Backup current binary
	timestamp := time.Now().Format("2006-01-02T15-04-05")
	backupPath := filepath.Join(u.backupDir, filepath.Base(u.binaryPath)+"."+timestamp)
	if err := os.Rename(u.binaryPath, backupPath); err != nil {
		os.Remove(tempFile)
		return "", fmt.Errorf("backup failed: %w", err)
	}

	// Swap new binary into place
	if err := os.Rename(tempFile, u.binaryPath); err != nil {
		// Try to restore backup
		_ = os.Rename(backupPath, u.binaryPath)
		return "", fmt.Errorf("swap failed: %w", err)
	}

	return backupPath, nil
}

// Rollback restores the previous binary from backup.
func (u *Updater) Rollback(backupPath string) error {
	if err := os.Rename(u.binaryPath, u.binaryPath+".failed"); err != nil {
		return fmt.Errorf("failed to move broken binary: %w", err)
	}

	if err := os.Rename(backupPath, u.binaryPath); err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
	}

	return nil
}

// downloadFile downloads a file from URL and saves it to path.
func downloadFile(ctx context.Context, url, path string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http %d", resp.StatusCode)
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}
