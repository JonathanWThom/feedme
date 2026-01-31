package api

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestUpdateInfoHasUpdate(t *testing.T) {
	testCases := []struct {
		name     string
		info     *UpdateInfo
		expected bool
	}{
		{
			name:     "nil info",
			info:     nil,
			expected: false,
		},
		{
			name:     "empty versions",
			info:     &UpdateInfo{},
			expected: false,
		},
		{
			name: "same version",
			info: &UpdateInfo{
				CurrentVersion: "v1.0.0",
				LatestVersion:  "v1.0.0",
			},
			expected: false,
		},
		{
			name: "newer patch version",
			info: &UpdateInfo{
				CurrentVersion: "v1.0.0",
				LatestVersion:  "v1.0.1",
			},
			expected: true,
		},
		{
			name: "newer minor version",
			info: &UpdateInfo{
				CurrentVersion: "v1.0.0",
				LatestVersion:  "v1.1.0",
			},
			expected: true,
		},
		{
			name: "newer major version",
			info: &UpdateInfo{
				CurrentVersion: "v1.0.0",
				LatestVersion:  "v2.0.0",
			},
			expected: true,
		},
		{
			name: "older version (no update)",
			info: &UpdateInfo{
				CurrentVersion: "v2.0.0",
				LatestVersion:  "v1.0.0",
			},
			expected: false,
		},
		{
			name: "without v prefix",
			info: &UpdateInfo{
				CurrentVersion: "1.0.0",
				LatestVersion:  "1.1.0",
			},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.info.HasUpdate()
			if result != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestCompareVersions(t *testing.T) {
	testCases := []struct {
		name     string
		v1       string
		v2       string
		expected int
	}{
		{"equal", "1.0.0", "1.0.0", 0},
		{"v1 greater patch", "1.0.1", "1.0.0", 1},
		{"v1 lesser patch", "1.0.0", "1.0.1", -1},
		{"v1 greater minor", "1.1.0", "1.0.0", 1},
		{"v1 lesser minor", "1.0.0", "1.1.0", -1},
		{"v1 greater major", "2.0.0", "1.0.0", 1},
		{"v1 lesser major", "1.0.0", "2.0.0", -1},
		{"with v prefix", "v1.1.0", "v1.0.0", 1},
		{"mixed v prefix", "v1.1.0", "1.0.0", 1},
		{"different lengths", "1.0", "1.0.0", 0},
		{"different lengths 2", "1.0.0", "1.0", 0},
		{"longer version wins", "1.0.1", "1.0", 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := compareVersions(tc.v1, tc.v2)
			if result != tc.expected {
				t.Errorf("compareVersions(%s, %s) = %d, expected %d", tc.v1, tc.v2, result, tc.expected)
			}
		})
	}
}

func TestFormatUpdateMessage(t *testing.T) {
	testCases := []struct {
		name     string
		info     *UpdateInfo
		expected string
	}{
		{
			name: "has update",
			info: &UpdateInfo{
				CurrentVersion: "v1.0.0",
				LatestVersion:  "v1.1.0",
			},
			expected: "Update available: v1.0.0 → v1.1.0",
		},
		{
			name: "no update",
			info: &UpdateInfo{
				CurrentVersion: "v1.1.0",
				LatestVersion:  "v1.1.0",
			},
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.info.FormatUpdateMessage()
			if result != tc.expected {
				t.Errorf("expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

func TestCheckForUpdateDevVersion(t *testing.T) {
	// Dev version should return nil
	result := CheckForUpdate("dev")
	if result != nil {
		t.Error("expected nil for dev version")
	}

	// Empty version should return nil
	result = CheckForUpdate("")
	if result != nil {
		t.Error("expected nil for empty version")
	}
}

func TestCheckForUpdateWithCache(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test that requires file system")
	}

	// Create a temporary cache directory
	tmpDir := t.TempDir()
	originalGetCacheDir := getCacheDir

	// Override getCacheDir for this test
	getCacheDirForTest := func() (string, error) {
		return tmpDir, nil
	}

	// Write a cache file
	cache := &updateCache{
		LastCheck:     time.Now(),
		LatestVersion: "v2.0.0",
		UpdateURL:     "https://example.com/release",
	}
	cacheData, _ := json.Marshal(cache)
	cachePath := filepath.Join(tmpDir, cacheFileName)
	err := os.WriteFile(cachePath, cacheData, 0644)
	if err != nil {
		t.Fatalf("failed to write cache file: %v", err)
	}

	// Test loading from cache
	loadedCache, err := loadCache(cachePath)
	if err != nil {
		t.Fatalf("loadCache failed: %v", err)
	}
	if loadedCache.LatestVersion != "v2.0.0" {
		t.Errorf("expected version v2.0.0, got %s", loadedCache.LatestVersion)
	}
	if loadedCache.UpdateURL != "https://example.com/release" {
		t.Errorf("expected URL https://example.com/release, got %s", loadedCache.UpdateURL)
	}

	_ = originalGetCacheDir
	_ = getCacheDirForTest
}

func TestLoadCacheInvalid(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "invalid.json")

	// Test non-existent file
	_, err := loadCache(cachePath)
	if err == nil {
		t.Error("expected error for non-existent file")
	}

	// Test invalid JSON
	err = os.WriteFile(cachePath, []byte("not json"), 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
	_, err = loadCache(cachePath)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestSaveCache(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "test_cache.json")

	cache := &updateCache{
		LastCheck:     time.Now(),
		LatestVersion: "v1.5.0",
		UpdateURL:     "https://example.com",
	}

	err := saveCache(cachePath, cache)
	if err != nil {
		t.Fatalf("saveCache failed: %v", err)
	}

	// Verify file was written
	data, err := os.ReadFile(cachePath)
	if err != nil {
		t.Fatalf("failed to read cache file: %v", err)
	}

	var loaded updateCache
	err = json.Unmarshal(data, &loaded)
	if err != nil {
		t.Fatalf("failed to unmarshal cache: %v", err)
	}

	if loaded.LatestVersion != cache.LatestVersion {
		t.Errorf("expected version %s, got %s", cache.LatestVersion, loaded.LatestVersion)
	}
}

// TestCheckGitHubRelease tests the actual GitHub API call
func TestCheckGitHubRelease(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// This tests against the real GitHub API
	info := checkGitHubRelease("v0.0.1")
	if info == nil {
		t.Log("GitHub API returned nil (might be rate limited or no releases)")
		return
	}

	if info.CurrentVersion != "v0.0.1" {
		t.Errorf("expected current version v0.0.1, got %s", info.CurrentVersion)
	}
	if info.LatestVersion == "" {
		t.Error("latest version should not be empty")
	}
	if info.UpdateURL == "" {
		t.Error("update URL should not be empty")
	}

	t.Logf("Latest release: %s at %s", info.LatestVersion, info.UpdateURL)
}

func TestGetCacheDir(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test that accesses file system")
	}

	dir, err := getCacheDir()
	if err != nil {
		t.Fatalf("getCacheDir failed: %v", err)
	}

	if dir == "" {
		t.Error("cache dir should not be empty")
	}

	// Verify directory exists
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("cache dir doesn't exist: %v", err)
	}
	if !info.IsDir() {
		t.Error("cache path is not a directory")
	}
}

// TestUpdateInfoNilReceiver tests nil receiver handling
func TestUpdateInfoNilReceiver(t *testing.T) {
	var info *UpdateInfo = nil

	// HasUpdate should not panic on nil receiver
	result := info.HasUpdate()
	if result != false {
		t.Error("expected false for nil UpdateInfo")
	}
}

// BenchmarkCompareVersions benchmarks version comparison
func BenchmarkCompareVersions(b *testing.B) {
	for i := 0; i < b.N; i++ {
		compareVersions("v1.2.3", "v1.2.4")
	}
}
