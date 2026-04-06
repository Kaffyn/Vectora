package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	vecos "github.com/Kaffyn/Vectora/internal/os"
)

type BinaryInfo struct {
	Name    string
	Version string
	URL     string
	Path    string
}

// Update all components or specified ones
func runUpdate(components []string, checkOnly bool) error {
	if checkOnly {
		fmt.Println("🔍 Checking for available updates...\n")
	} else {
		fmt.Println("🔍 Checking for updates and installing...\n")
	}

	// If no components specified, update all
	if len(components) == 0 {
		components = []string{"daemon", "tui", "lpm", "mpm", "setup"}
	}

	systemManager, _ := vecos.NewManager()
	var installDir string
	if systemManager != nil {
		dir, _ := systemManager.GetInstallDir()
		installDir = dir
	}

	if installDir == "" {
		exe, _ := os.Executable()
		installDir = filepath.Dir(exe)
	}

	for _, component := range components {
		if err := updateComponent(component, installDir, checkOnly); err != nil {
			fmt.Printf("❌ Failed to update %s: %v\n", component, err)
			continue
		}
	}

	if checkOnly {
		fmt.Println("\n✅ Update check completed!")
	} else {
		fmt.Println("\n✅ Update process completed!")
	}
	return nil
}

func updateComponent(component, installDir string, checkOnly bool) error {
	var binaryName string
	var suffix string

	if runtime.GOOS == "windows" {
		suffix = ".exe"
	}

	switch component {
	case "daemon":
		binaryName = "vectora" + suffix
	case "tui", "cli", "desktop":
		binaryName = "vectora-tui" + suffix
	case "lpm":
		binaryName = "lpm" + suffix
	case "mpm":
		binaryName = "mpm" + suffix
	case "setup", "installer":
		binaryName = "vectora-setup" + suffix
	default:
		return fmt.Errorf("unknown component: %s", component)
	}

	binaryPath := filepath.Join(installDir, binaryName)
	currentVersion, _ := getBinaryVersion(binaryPath)

	fmt.Printf("📦 %s\n", component)
	fmt.Printf("   Current version: %s\n", currentVersion)

	// Check for updates from releases
	latestVersion, downloadURL := getLatestRelease(component)
	if latestVersion == "" {
		fmt.Printf("   Status: No updates available (offline mode)\n")
		return nil
	}

	if latestVersion == currentVersion {
		fmt.Printf("   Status: ✓ Already up to date\n")
		return nil
	}

	fmt.Printf("   Latest version: %s\n", latestVersion)

	if checkOnly {
		fmt.Printf("   Status: Update available (use 'vectora update %s' to install)\n", component)
		return nil
	}

	fmt.Printf("   Status: ⬇️  Downloading update...\n")

	// Download and update
	if err := downloadAndReplace(downloadURL, binaryPath, binaryName); err != nil {
		return err
	}

	fmt.Printf("   Status: ✅ Updated to %s\n", latestVersion)
	return nil
}

func getBinaryVersion(binaryPath string) (string, error) {
	if _, err := os.Stat(binaryPath); err != nil {
		return "not installed", nil
	}

	cmd := exec.Command(binaryPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "unknown", nil
	}

	// Parse version from output - extract first occurrence of v?.?.?
	versionStr := strings.TrimSpace(string(output))

	// Remove version prefix if present
	if strings.Contains(versionStr, "v") {
		parts := strings.Fields(versionStr)
		for _, part := range parts {
			if strings.HasPrefix(part, "v") && len(part) > 1 {
				return part[1:], nil
			}
		}
	}

	// Try to extract just the numeric version
	if len(versionStr) > 0 {
		return versionStr, nil
	}

	return "unknown", nil
}

type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name        string `json:"name"`
		DownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func getLatestRelease(component string) (version, url string) {
	// Query GitHub API for latest release
	apiURL := "https://api.github.com/repos/Kaffyn/Vectora/releases/latest"

	resp, err := http.Get(apiURL)
	if err != nil {
		// Offline mode - use fallback versions
		return getFallbackVersion(component), ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return getFallbackVersion(component), ""
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return getFallbackVersion(component), ""
	}

	version = release.TagName
	if strings.HasPrefix(version, "v") {
		version = version[1:]
	}

	// Find the correct asset for this component and platform
	binaryName := getBinaryName(component)
	for _, asset := range release.Assets {
		if strings.Contains(asset.Name, binaryName) && strings.Contains(asset.Name, runtime.GOOS) {
			url = asset.DownloadURL
			break
		}
	}

	return version, url
}

func getFallbackVersion(component string) string {
	// Fallback versions for offline mode
	versions := map[string]string{
		"daemon":  "0.1.0",
		"tui":     "0.1.0",
		"lpm":     "0.1.0",
		"mpm":     "0.1.0",
		"setup":   "0.1.0",
	}
	return versions[component]
}

func getBinaryName(component string) string {
	switch component {
	case "daemon":
		return "vectora"
	case "tui", "cli", "desktop":
		return "vectora-tui"
	case "lpm":
		return "lpm"
	case "mpm":
		return "mpm"
	case "setup", "installer":
		return "vectora-setup"
	default:
		return component
	}
}

func downloadAndReplace(downloadURL, targetPath, binaryName string) error {
	if downloadURL == "" {
		return fmt.Errorf("no download URL available")
	}

	// Create backup
	backupPath := targetPath + ".backup"
	if _, err := os.Stat(targetPath); err == nil {
		if err := copyFile(targetPath, backupPath); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
	}

	// Download new binary
	tempFile := targetPath + ".tmp"
	if err := downloadFile(downloadURL, tempFile); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	// Stop daemon if updating daemon itself
	if binaryName == "vectora" || binaryName == "vectora.exe" {
		stopDaemonGracefully()
		time.Sleep(1 * time.Second)
	}

	// Replace old binary with new one
	if err := os.Remove(targetPath); err != nil {
		os.Remove(tempFile)
		os.Rename(backupPath, targetPath)
		return fmt.Errorf("failed to remove old binary: %w", err)
	}

	if err := os.Rename(tempFile, targetPath); err != nil {
		os.Rename(backupPath, targetPath)
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	// Set executable permissions
	os.Chmod(targetPath, 0755)

	// Clean up backup
	os.Remove(backupPath)

	return nil
}

func downloadFile(url, targetPath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	out, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

func stopDaemonGracefully() {
	// Try to stop daemon gracefully via IPC
	cmd := exec.Command("vectora", "stop")
	_ = cmd.Run()
}
