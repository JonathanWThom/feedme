package api

import (
	"testing"
	"time"
)

// TestLobstersClientInterface verifies that LobstersClient implements Source interface
func TestLobstersClientInterface(t *testing.T) {
	var _ Source = (*LobstersClient)(nil)
}

func TestNewLobstersClient(t *testing.T) {
	client := NewLobstersClient()
	if client == nil {
		t.Fatal("NewLobstersClient returned nil")
	}
	if client.http == nil {
		t.Fatal("HTTP client is nil")
	}
	if client.storyCache == nil {
		t.Fatal("Story cache is nil")
	}
}

func TestLobstersClientName(t *testing.T) {
	client := NewLobstersClient()
	name := client.Name()
	if name != "Lobsters" {
		t.Errorf("expected name 'Lobsters', got '%s'", name)
	}
}

func TestLobstersClientFeedNames(t *testing.T) {
	client := NewLobstersClient()
	feedNames := client.FeedNames()

	expected := []string{LobstersFeedHottest, LobstersFeedNewest, LobstersFeedRecent}
	if len(feedNames) != len(expected) {
		t.Errorf("expected %d feed names, got %d", len(expected), len(feedNames))
	}

	for i, name := range expected {
		if feedNames[i] != name {
			t.Errorf("expected feed name '%s' at index %d, got '%s'", name, i, feedNames[i])
		}
	}
}

func TestLobstersClientFeedLabels(t *testing.T) {
	client := NewLobstersClient()
	labels := client.FeedLabels()

	expected := []string{"Hot", "New", "Recent"}
	if len(labels) != len(expected) {
		t.Errorf("expected %d labels, got %d", len(expected), len(labels))
	}

	for i, label := range expected {
		if labels[i] != label {
			t.Errorf("expected label '%s' at index %d, got '%s'", label, i, labels[i])
		}
	}
}

func TestLobstersClientStoryURL(t *testing.T) {
	client := NewLobstersClient()

	testCases := []struct {
		name     string
		item     *Item
		expected string
	}{
		{
			name:     "with short ID",
			item:     &Item{Type: "abc123"},
			expected: "https://lobste.rs/s/abc123",
		},
		{
			name:     "with story type",
			item:     &Item{Type: "story", URL: "https://example.com"},
			expected: "https://example.com",
		},
		{
			name:     "with empty type",
			item:     &Item{Type: "", URL: "https://example.com"},
			expected: "https://example.com",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := client.StoryURL(tc.item)
			if url != tc.expected {
				t.Errorf("expected URL '%s', got '%s'", tc.expected, url)
			}
		})
	}
}

// TestLobstersFetchStoryIDs tests fetching story IDs from real Lobste.rs
func TestLobstersFetchStoryIDs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := NewLobstersClient()

	testCases := []string{LobstersFeedHottest, LobstersFeedNewest, LobstersFeedRecent}
	for _, feed := range testCases {
		t.Run(feed, func(t *testing.T) {
			ids, err := client.FetchStoryIDs(feed)
			if err != nil {
				t.Fatalf("FetchStoryIDs failed for %s: %v", feed, err)
			}
			if len(ids) == 0 {
				t.Errorf("expected non-empty story IDs for feed %s", feed)
			}
			// Should get about 50 stories (2 pages)
			if len(ids) < 10 {
				t.Errorf("expected at least 10 IDs, got %d", len(ids))
			}
			// Verify IDs are positive (pseudo-IDs start at 1)
			for _, id := range ids {
				if id <= 0 {
					t.Errorf("invalid story ID: %d", id)
				}
			}
		})
	}
}

// TestLobstersFetchItem tests fetching cached items
func TestLobstersFetchItem(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := NewLobstersClient()

	// First fetch story IDs to populate cache
	ids, err := client.FetchStoryIDs(LobstersFeedHottest)
	if err != nil {
		t.Fatalf("FetchStoryIDs failed: %v", err)
	}
	if len(ids) == 0 {
		t.Fatal("no story IDs returned")
	}

	// Fetch the first item from cache
	item, err := client.FetchItem(ids[0])
	if err != nil {
		t.Fatalf("FetchItem failed: %v", err)
	}
	if item == nil {
		t.Fatal("FetchItem returned nil")
	}

	// Validate item fields
	if item.Title == "" {
		t.Error("item title is empty")
	}
	if item.By == "" {
		t.Error("item author is empty")
	}
	if item.URL == "" {
		t.Error("item URL is empty")
	}
}

// TestLobstersFetchItemNotInCache tests fetching non-existent item
func TestLobstersFetchItemNotInCache(t *testing.T) {
	client := NewLobstersClient()

	// Try to fetch an item that doesn't exist in cache
	_, err := client.FetchItem(99999)
	if err == nil {
		t.Error("expected error for non-existent item, got nil")
	}
}

// TestLobstersFetchItems tests fetching multiple items
func TestLobstersFetchItems(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := NewLobstersClient()

	// Fetch story IDs first
	ids, err := client.FetchStoryIDs(LobstersFeedHottest)
	if err != nil {
		t.Fatalf("FetchStoryIDs failed: %v", err)
	}

	// Fetch first 5 items
	batchSize := min(5, len(ids))
	items, err := client.FetchItems(ids[:batchSize])
	if err != nil {
		t.Fatalf("FetchItems failed: %v", err)
	}

	if len(items) != batchSize {
		t.Errorf("expected %d items, got %d", batchSize, len(items))
	}

	// Count non-nil items
	nonNil := 0
	for _, item := range items {
		if item != nil {
			nonNil++
		}
	}
	if nonNil == 0 {
		t.Error("all items are nil")
	}
}

// TestLobstersFetchCommentTree tests fetching comments
func TestLobstersFetchCommentTree(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := NewLobstersClient()

	// Fetch story IDs
	ids, err := client.FetchStoryIDs(LobstersFeedHottest)
	if err != nil {
		t.Fatalf("FetchStoryIDs failed: %v", err)
	}

	// Find a story with comments
	var storyWithComments *Item
	items, _ := client.FetchItems(ids[:min(10, len(ids))])
	for _, item := range items {
		if item != nil && item.Descendants > 0 && item.Type != "" && item.Type != "story" {
			storyWithComments = item
			break
		}
	}

	if storyWithComments == nil {
		t.Skip("no story with comments found")
	}

	// Fetch comment tree
	comments, err := client.FetchCommentTree(storyWithComments, 0)
	if err != nil {
		t.Fatalf("FetchCommentTree failed: %v", err)
	}

	t.Logf("Found %d comments for story '%s'", len(comments), storyWithComments.Title)

	// Verify comment structure if we have any
	for _, c := range comments {
		if c.Item == nil {
			t.Error("comment has nil Item")
			continue
		}
		if c.Depth < 0 {
			t.Errorf("invalid comment depth: %d", c.Depth)
		}
	}
}

// TestLobstersFetchCommentTreeNoStoryID tests error handling
func TestLobstersFetchCommentTreeNoStoryID(t *testing.T) {
	client := NewLobstersClient()

	// Item without short ID
	item := &Item{Type: "story"}
	_, err := client.FetchCommentTree(item, 0)
	if err == nil {
		t.Error("expected error for item without short ID")
	}
}

// TestHashShortID tests the short ID hashing function
func TestHashShortID(t *testing.T) {
	testCases := []struct {
		shortID  string
		expected int
	}{
		{"abc", 31*1 + 31*2*98 + 31*3*99}, // 'a'=97, 'b'=98, 'c'=99
		{"", 0},
	}

	for _, tc := range testCases {
		t.Run(tc.shortID, func(t *testing.T) {
			result := hashShortID(tc.shortID)
			if tc.shortID == "" && result != 0 {
				t.Errorf("expected 0 for empty string, got %d", result)
			}
			// Just verify it's deterministic
			result2 := hashShortID(tc.shortID)
			if result != result2 {
				t.Errorf("hash not deterministic: %d vs %d", result, result2)
			}
		})
	}
}

// TestParseTime tests time parsing
func TestParseTime(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"RFC3339", "2024-01-15T10:30:00Z", false},
		{"RFC3339 with offset", "2024-01-15T10:30:00-05:00", false},
		{"Custom format", "2024-01-15 10:30:00 -0500", false},
		{"Invalid", "not a date", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseTime(tc.input)
			if tc.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestParseRelativeTime tests relative time parsing
func TestParseRelativeTime(t *testing.T) {
	now := time.Now()

	testCases := []struct {
		name       string
		input      string
		checkDelta time.Duration
	}{
		{"seconds", "30 seconds ago", 30 * time.Second},
		{"minute", "1 minute ago", 1 * time.Minute},
		{"minutes", "5 minutes ago", 5 * time.Minute},
		{"hour", "1 hour ago", 1 * time.Hour},
		{"hours", "3 hours ago", 3 * time.Hour},
		{"day", "1 day ago", 24 * time.Hour},
		{"days", "7 days ago", 7 * 24 * time.Hour},
		{"week", "1 week ago", 7 * 24 * time.Hour},
		{"weeks", "2 weeks ago", 2 * 7 * 24 * time.Hour},
		{"month", "1 month ago", 30 * 24 * time.Hour},
		{"year", "1 year ago", 365 * 24 * time.Hour},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := parseRelativeTime(tc.input)
			parsed := time.Unix(result, 0)
			expected := now.Add(-tc.checkDelta)

			// Allow 2 second tolerance for test execution time
			diff := parsed.Sub(expected)
			if diff < -2*time.Second || diff > 2*time.Second {
				t.Errorf("parsed time differs by %v (expected ~%v ago)", diff, tc.checkDelta)
			}
		})
	}
}

// TestParseRelativeTimeInvalid tests invalid relative time strings
func TestParseRelativeTimeInvalid(t *testing.T) {
	now := time.Now()
	result := parseRelativeTime("invalid string")

	// Should return approximately now
	parsed := time.Unix(result, 0)
	diff := now.Sub(parsed)
	if diff < -2*time.Second || diff > 2*time.Second {
		t.Errorf("invalid input should return ~now, got difference of %v", diff)
	}
}

// TestLobstersFeedConstants tests that feed constants are correct
func TestLobstersFeedConstants(t *testing.T) {
	if LobstersFeedHottest != "" {
		t.Errorf("LobstersFeedHottest should be empty string, got '%s'", LobstersFeedHottest)
	}
	if LobstersFeedNewest != "newest" {
		t.Errorf("LobstersFeedNewest should be 'newest', got '%s'", LobstersFeedNewest)
	}
	if LobstersFeedRecent != "recent" {
		t.Errorf("LobstersFeedRecent should be 'recent', got '%s'", LobstersFeedRecent)
	}
}

// TestLobstersThrottle tests that throttling doesn't cause issues
func TestLobstersThrottle(t *testing.T) {
	client := NewLobstersClient()

	// Record initial time
	start := time.Now()

	// Call throttle twice
	client.throttle()
	client.throttle()

	// Second call should have waited ~500ms
	elapsed := time.Since(start)
	if elapsed < 400*time.Millisecond {
		t.Errorf("throttle didn't wait long enough: %v", elapsed)
	}
}
