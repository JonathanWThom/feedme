# Adding a New Provider to FeedMe

This guide explains how to add a new news source provider to FeedMe. The application uses a clean provider interface pattern that makes it straightforward to add new sources.

## Architecture Overview

FeedMe uses the `Source` interface defined in `api/source.go` to abstract different news sources. Each provider must implement this interface to integrate with the application.

```
feedme/
├── api/
│   ├── source.go      # Source interface definition
│   ├── client.go      # Hacker News provider
│   ├── lobsters.go    # Lobsters provider
│   ├── reddit.go      # Reddit provider
│   └── <your_provider>.go  # Your new provider
├── ui/
│   └── model.go       # UI integration
└── main.go            # CLI entry point
```

## Step 1: Understand the Source Interface

The `Source` interface in `api/source.go` defines what every provider must implement:

```go
type Source interface {
    // Name returns the display name of the source
    Name() string

    // FeedNames returns the available feed names for this source
    // Example: ["topstories", "newstories", "beststories"]
    FeedNames() []string

    // FeedLabels returns the display labels for feeds (for UI tabs)
    // Example: ["Top", "New", "Best"]
    FeedLabels() []string

    // FetchStoryIDs fetches the list of story IDs for a given feed
    // For sources without IDs, return page-based pseudo-IDs
    FetchStoryIDs(feed string) ([]int, error)

    // FetchItem fetches a single item by ID
    FetchItem(id int) (*Item, error)

    // FetchItems fetches multiple items by ID
    FetchItems(ids []int) ([]*Item, error)

    // FetchCommentTree fetches the comment tree for a story
    FetchCommentTree(item *Item, maxDepth int) ([]*Comment, error)

    // StoryURL returns the URL for viewing a story on the source's website
    StoryURL(item *Item) string
}
```

## Step 2: Create Your Provider File

Create a new file `api/<your_provider>.go`. Here's a template:

```go
package api

import (
    "fmt"
    "net/http"
    "sync"
    "time"
)

// Feed types for your provider
const (
    YourFeedHot    = "hot"
    YourFeedNew    = "new"
    // Add more feeds as needed
)

var YourFeedNames = []string{YourFeedHot, YourFeedNew}
var YourFeedLabels = []string{"Hot", "New"}

// YourClient fetches data from your source
type YourClient struct {
    http        *http.Client
    storyCache  map[int]*Item    // Cache for items (if needed)
    cacheMu     sync.RWMutex     // Mutex for cache access
    lastRequest time.Time        // For rate limiting
    requestMu   sync.Mutex       // Mutex for rate limiting
}

// NewYourClient creates a new client for your source
func NewYourClient() *YourClient {
    return &YourClient{
        http: &http.Client{
            Timeout: 15 * time.Second,
        },
        storyCache: make(map[int]*Item),
    }
}

// throttle ensures we don't make requests too quickly
func (c *YourClient) throttle() {
    c.requestMu.Lock()
    defer c.requestMu.Unlock()

    minDelay := 500 * time.Millisecond // Adjust based on API limits
    elapsed := time.Since(c.lastRequest)
    if elapsed < minDelay {
        time.Sleep(minDelay - elapsed)
    }
    c.lastRequest = time.Now()
}

// Name returns the display name of the source
func (c *YourClient) Name() string {
    return "YourSource"
}

// FeedNames returns the available feed names
func (c *YourClient) FeedNames() []string {
    return YourFeedNames
}

// FeedLabels returns the display labels for feeds
func (c *YourClient) FeedLabels() []string {
    return YourFeedLabels
}

// StoryURL returns the URL for viewing a story on your source
func (c *YourClient) StoryURL(item *Item) string {
    return fmt.Sprintf("https://your-source.com/item/%d", item.ID)
}

// FetchStoryIDs fetches story IDs for a feed
func (c *YourClient) FetchStoryIDs(feed string) ([]int, error) {
    c.throttle()

    // Option A: If your API returns numeric IDs directly
    // url := fmt.Sprintf("https://api.your-source.com/%s.json", feed)
    // ... fetch and return IDs

    // Option B: If you need to scrape/cache (like Lobsters)
    // Fetch stories, cache them, return pseudo-IDs
    stories, err := c.fetchStories(feed)
    if err != nil {
        return nil, err
    }

    ids := make([]int, len(stories))
    c.cacheMu.Lock()
    c.storyCache = make(map[int]*Item) // Clear old cache
    for i, story := range stories {
        id := i + 1 // 1-indexed pseudo-IDs
        c.storyCache[id] = story
        ids[i] = id
    }
    c.cacheMu.Unlock()

    return ids, nil
}

// fetchStories fetches stories from your source
func (c *YourClient) fetchStories(feed string) ([]*Item, error) {
    // Implement your fetching logic here
    // Return a slice of *Item
    return nil, nil
}

// FetchItem fetches a single item by ID
func (c *YourClient) FetchItem(id int) (*Item, error) {
    // If using cache:
    c.cacheMu.RLock()
    item, ok := c.storyCache[id]
    c.cacheMu.RUnlock()

    if !ok {
        return nil, fmt.Errorf("item %d not found in cache", id)
    }
    return item, nil
}

// FetchItems fetches multiple items by ID
func (c *YourClient) FetchItems(ids []int) ([]*Item, error) {
    items := make([]*Item, len(ids))
    c.cacheMu.RLock()
    for i, id := range ids {
        if item, ok := c.storyCache[id]; ok {
            items[i] = item
        }
    }
    c.cacheMu.RUnlock()
    return items, nil
}

// FetchCommentTree fetches comments for a story
func (c *YourClient) FetchCommentTree(item *Item, maxDepth int) ([]*Comment, error) {
    c.throttle()
    // Implement your comment fetching logic
    return nil, nil
}
```

## Step 3: Work with the Item and Comment Types

The `Item` struct (defined in `api/client.go`) is shared across all providers:

```go
type Item struct {
    ID          int    `json:"id"`
    Type        string `json:"type"`    // Can store additional data (e.g., permalink)
    By          string `json:"by"`      // Author
    Time        int64  `json:"time"`    // Unix timestamp
    Text        string `json:"text"`    // Body text or tags
    URL         string `json:"url"`     // Link URL
    Title       string `json:"title"`
    Score       int    `json:"score"`   // Points/upvotes
    Kids        []int  `json:"kids"`    // Child IDs (for HN-style APIs)
    Descendants int    `json:"descendants"` // Comment count
    Deleted     bool   `json:"deleted"`
    Dead        bool   `json:"dead"`
}
```

The `Comment` struct represents a comment with depth information:

```go
type Comment struct {
    *Item
    Depth    int
    Children []*Comment
}
```

### Tips for Mapping Data

- **ID**: Use actual numeric IDs if available, or generate pseudo-IDs (1-indexed)
- **Type**: Use for storing provider-specific data (e.g., short IDs, permalinks)
- **Text**: Can store tags, flair, or other metadata (displayed after score)
- **Score**: Points, upvotes, or other engagement metric
- **Descendants**: Total comment count (shown in UI)

## Step 4: Register Your Provider in main.go

Add your provider to the source selection in `main.go`:

```go
switch {
case sourceLower == "hn" || sourceLower == "hackernews" || sourceLower == "hacker-news":
    source = api.NewClient()
case sourceLower == "lobsters" || sourceLower == "lobste.rs" || sourceLower == "l":
    source = api.NewLobstersClient()
case strings.HasPrefix(sourceLower, "r/") || strings.HasPrefix(sourceLower, "/r/"):
    source = api.NewRedditClient(sourceFlag)
// Add your provider:
case sourceLower == "yoursource" || sourceLower == "ys":
    source = api.NewYourClient()
default:
    // Update error message
    fmt.Fprintf(os.Stderr, "Valid sources: hn, lobsters, r/subreddit, yoursource\n")
}
```

## Step 5: Add to Source Picker (Optional)

To add your provider to the interactive source picker, update `ui/model.go`:

```go
// Source picker options
var sourceOptions = []string{"Hacker News", "Lobste.rs", "Reddit", "YourSource"}

// In handleSourcePickerInput:
case 3: // YourSource
    m.source = api.NewYourClient()
    m.view = StoriesView
    // ... rest of initialization
```

## Step 6: Write Tests

Create `api/<your_provider>_test.go` with tests for your provider:

```go
package api

import (
    "testing"
)

// TestYourClientInterface verifies interface implementation
func TestYourClientInterface(t *testing.T) {
    var _ Source = (*YourClient)(nil)
}

func TestNewYourClient(t *testing.T) {
    client := NewYourClient()
    if client == nil {
        t.Fatal("NewYourClient returned nil")
    }
}

func TestYourClientName(t *testing.T) {
    client := NewYourClient()
    if client.Name() != "YourSource" {
        t.Errorf("unexpected name: %s", client.Name())
    }
}

// Integration tests (use -short to skip)
func TestYourFetchStoryIDs(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    client := NewYourClient()
    ids, err := client.FetchStoryIDs(YourFeedHot)
    if err != nil {
        t.Fatalf("FetchStoryIDs failed: %v", err)
    }
    if len(ids) == 0 {
        t.Error("expected non-empty IDs")
    }
}

// Add more tests...
```

## Best Practices

### Rate Limiting

Always implement rate limiting to be a good API citizen:

```go
func (c *YourClient) throttle() {
    c.requestMu.Lock()
    defer c.requestMu.Unlock()

    minDelay := 500 * time.Millisecond
    elapsed := time.Since(c.lastRequest)
    if elapsed < minDelay {
        time.Sleep(minDelay - elapsed)
    }
    c.lastRequest = time.Now()
}
```

### Error Handling

Handle rate limit responses (HTTP 429):

```go
if resp.StatusCode == 429 {
    resp.Body.Close()
    time.Sleep(2 * time.Second)
    // Retry once
}
```

### User-Agent

Set a descriptive User-Agent header:

```go
req.Header.Set("User-Agent", "feedme:v1.0 (terminal news reader)")
```

### Caching Strategy

Choose the right caching approach:

1. **No cache** (like HN): API returns IDs, fetch items on demand
2. **Full cache** (like Lobsters/Reddit): Scrape/fetch all at once, store in memory

### HTML Parsing

For scraping, use `github.com/PuerkitoBio/goquery`:

```go
import "github.com/PuerkitoBio/goquery"

doc, err := goquery.NewDocumentFromReader(resp.Body)
if err != nil {
    return nil, err
}

doc.Find("article.story").Each(func(i int, s *goquery.Selection) {
    title := s.Find("h2.title").Text()
    // ...
})
```

### JSON APIs

For JSON APIs, define response structs:

```go
type apiResponse struct {
    Data struct {
        Stories []struct {
            ID    string `json:"id"`
            Title string `json:"title"`
        } `json:"stories"`
    } `json:"data"`
}

var result apiResponse
json.NewDecoder(resp.Body).Decode(&result)
```

## Example: Complete Minimal Provider

Here's a minimal but complete provider implementation:

```go
package api

import (
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

type MinimalClient struct {
    http *http.Client
}

func NewMinimalClient() *MinimalClient {
    return &MinimalClient{
        http: &http.Client{Timeout: 10 * time.Second},
    }
}

func (c *MinimalClient) Name() string           { return "Minimal" }
func (c *MinimalClient) FeedNames() []string    { return []string{"feed"} }
func (c *MinimalClient) FeedLabels() []string   { return []string{"Feed"} }
func (c *MinimalClient) StoryURL(item *Item) string {
    return fmt.Sprintf("https://example.com/%d", item.ID)
}

func (c *MinimalClient) FetchStoryIDs(feed string) ([]int, error) {
    // Return some IDs
    return []int{1, 2, 3, 4, 5}, nil
}

func (c *MinimalClient) FetchItem(id int) (*Item, error) {
    return &Item{
        ID:    id,
        Title: fmt.Sprintf("Story %d", id),
        By:    "author",
        Score: 100,
        Time:  time.Now().Unix(),
    }, nil
}

func (c *MinimalClient) FetchItems(ids []int) ([]*Item, error) {
    items := make([]*Item, len(ids))
    for i, id := range ids {
        items[i], _ = c.FetchItem(id)
    }
    return items, nil
}

func (c *MinimalClient) FetchCommentTree(item *Item, maxDepth int) ([]*Comment, error) {
    return []*Comment{
        {Item: &Item{By: "user", Text: "A comment", Time: time.Now().Unix()}, Depth: 0},
    }, nil
}
```

## Testing Your Provider

Run unit tests:
```bash
go test ./api -v -run YourProvider
```

Run integration tests (hits real APIs):
```bash
go test ./api -v -run Integration
```

Test with the CLI:
```bash
go run . -s yoursource
```

## Checklist

Before submitting your provider:

- [ ] Implements all `Source` interface methods
- [ ] Has proper rate limiting
- [ ] Sets appropriate User-Agent
- [ ] Handles HTTP errors gracefully
- [ ] Has unit tests
- [ ] Has integration tests
- [ ] Registered in `main.go`
- [ ] Added to source picker (optional)
- [ ] Documentation updated

## Getting Help

If you need help or have questions:

1. Look at existing providers for reference:
   - `api/client.go` - JSON API with concurrent fetching
   - `api/lobsters.go` - HTML scraping with caching
   - `api/reddit.go` - JSON API with nested comment parsing

2. Open an issue on GitHub for guidance

3. Submit a draft PR for early feedback
