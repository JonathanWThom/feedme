package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

// updateCache stores the last check time and result
type updateCache struct {
	LastCheck     time.Time `json:"last_check"`
	LatestVersion string    `json:"latest_version"`
	UpdateURL     string    `json:"update_url"`
}

// CheckForUpdate checks if a newer version is available
// It uses a cached result if checked within the last 24 hours
func CheckForUpdate(currentVersion string) *UpdateInfo {
	if currentVersion == "" || currentVersion == "dev" {
		return nil
	}

	cacheDir, err := getCacheDir()
	if err != nil {
		return checkGitHubRelease(currentVersion)
	}

	cachePath := filepath.Join(cacheDir, cacheFileName)
	cache, err := loadCache(cachePath)
	if err == nil && time.Since(cache.LastCheck) < checkInterval {
		return &UpdateInfo{
			CurrentVersion: currentVersion,
			LatestVersion:  cache.LatestVersion,
			UpdateURL:      cache.UpdateURL,
		}
	}

	info := checkGitHubRelease(currentVersion)
	if info != nil {
		saveCache(cachePath, &updateCache{
			LastCheck:     time.Now(),
			LatestVersion: info.LatestVersion,
			UpdateURL:     info.UpdateURL,
		})
	}

	return info
}

func checkGitHubRelease(currentVersion string) *UpdateInfo {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(githubReleaseURL)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil
	}

	return &UpdateInfo{
		CurrentVersion: currentVersion,
		LatestVersion:  release.TagName,
		UpdateURL:      release.HTMLURL,
	}
}

func getCacheDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	cacheDir := filepath.Join(configDir, "feedme")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", err
	}

	return cacheDir, nil
}

func loadCache(path string) (*updateCache, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cache updateCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}

	return &cache, nil
}

func saveCache(path string, cache *updateCache) error {
	data, err := json.Marshal(cache)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// compareVersions compares two semantic version strings
// Returns: 1 if v1 > v2, -1 if v1 < v2, 0 if equal
func compareVersions(v1, v2 string) int {
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")

	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var n1, n2 int
		if i < len(parts1) {
			n1, _ = strconv.Atoi(parts1[i])
		}
		if i < len(parts2) {
			n2, _ = strconv.Atoi(parts2[i])
		}

		if n1 > n2 {
			return 1
		}
		if n1 < n2 {
			return -1
		}
	}

	return 0
}

// FormatUpdateMessage returns a formatted message for the status bar
func (u *UpdateInfo) FormatUpdateMessage() string {
	if !u.HasUpdate() {
		return ""
	}
	return fmt.Sprintf("Update available: %s → %s", u.CurrentVersion, u.LatestVersion)
}
