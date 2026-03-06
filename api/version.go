package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"time"
)

const (
	githubReleaseURL = "https://api.github.com/repos/JonathanWThom/feedme/releases/latest"
	cacheFileName    = "update_check.json"
	checkInterval    = 24 * time.Hour
)

// UpdateInfo contains information about an available update
type UpdateInfo struct {
	CurrentVersion string
	LatestVersion  string
	UpdateURL      string
}

// HasUpdate returns true if a newer version is available
func (u *UpdateInfo) HasUpdate() bool {
	return u != nil && u.LatestVersion != "" && u.LatestVersion != u.CurrentVersion &&
		compareVersions(u.LatestVersion, u.CurrentVersion) > 0
}

// githubRelease represents the GitHub API response for a release
type githubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

// CheckForUpdate checks if a newer version is available
// It uses a cached result if checked within the last 24 hours
func CheckForUpdate(currentVersion string) *UpdateInfo {
	if currentVersion == "" || currentVersion == "dev" {
		return nil
	}
	cachePath, err := updateCachePath()
	if err != nil {
		return fetchRelease(currentVersion)
	}
	if info := checkCachedUpdate(cachePath, currentVersion); info != nil {
		return info
	}
	return fetchAndCache(cachePath, currentVersion)
}

func updateCachePath() (string, error) {
	cacheDir, err := getCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cacheDir, cacheFileName), nil
}

func checkCachedUpdate(cachePath, currentVersion string) *UpdateInfo {
	cache, err := loadCache(cachePath)
	if err != nil || time.Since(cache.LastCheck) >= checkInterval {
		return nil
	}
	return &UpdateInfo{
		CurrentVersion: currentVersion,
		LatestVersion:  cache.LatestVersion,
		UpdateURL:      cache.UpdateURL,
	}
}

func fetchAndCache(cachePath, currentVersion string) *UpdateInfo {
	info := fetchRelease(currentVersion)
	if info != nil {
		saveCache(cachePath, &updateCache{
			LastCheck:     time.Now(),
			LatestVersion: info.LatestVersion,
			UpdateURL:     info.UpdateURL,
		})
	}
	return info
}

func fetchRelease(currentVersion string) *UpdateInfo {
	release, err := fetchLatestRelease()
	if err != nil {
		return nil
	}
	return &UpdateInfo{
		CurrentVersion: currentVersion,
		LatestVersion:  release.TagName,
		UpdateURL:      release.HTMLURL,
	}
}

func fetchLatestRelease() (*githubRelease, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(githubReleaseURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	var release githubRelease
	err = json.NewDecoder(resp.Body).Decode(&release)
	return &release, err
}

// FormatUpdateMessage returns a formatted message for the status bar
func (u *UpdateInfo) FormatUpdateMessage() string {
	if !u.HasUpdate() {
		return ""
	}
	return fmt.Sprintf("Update available: %s → %s", u.CurrentVersion, u.LatestVersion)
}
