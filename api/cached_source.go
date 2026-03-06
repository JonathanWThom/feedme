package api

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// CachedSource provides throttling and in-memory story caching
// shared by sources that fetch stories in bulk (Lobsters, Reddit).
type CachedSource struct {
	storyCache  map[int]*Item
	cacheMu     sync.RWMutex
	lastRequest time.Time
	requestMu   sync.Mutex
	minDelay    time.Duration
}

// NewCachedSource creates a CachedSource with the given throttle delay.
func NewCachedSource(minDelay time.Duration) CachedSource {
	return CachedSource{
		storyCache: make(map[int]*Item),
		minDelay:   minDelay,
	}
}

// Throttle ensures we don't make requests too quickly.
func (c *CachedSource) Throttle() {
	c.requestMu.Lock()
	defer c.requestMu.Unlock()

	elapsed := time.Since(c.lastRequest)
	if elapsed < c.minDelay {
		time.Sleep(c.minDelay - elapsed)
	}
	c.lastRequest = time.Now()
}

// StoreItems clears the cache and stores items with 1-indexed pseudo-IDs.
// Returns the generated IDs.
func (c *CachedSource) StoreItems(items []*Item) []int {
	ids := make([]int, len(items))
	c.cacheMu.Lock()
	c.storyCache = make(map[int]*Item)
	for i, item := range items {
		id := i + 1
		c.storyCache[id] = item
		ids[i] = id
	}
	c.cacheMu.Unlock()
	return ids
}

// FetchItem fetches a cached item by pseudo-ID.
func (c *CachedSource) FetchItem(id int) (*Item, error) {
	c.cacheMu.RLock()
	item, ok := c.storyCache[id]
	c.cacheMu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("item %d not found in cache", id)
	}
	return item, nil
}

// FetchItems fetches multiple cached items by pseudo-ID.
func (c *CachedSource) FetchItems(ids []int) ([]*Item, error) {
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

// doWithRetry makes an HTTP request, retrying once on 429 rate limiting.
func doWithRetry(client *http.Client, url, userAgent string, cs *CachedSource) (*http.Response, error) {
	resp, err := doRequest(client, url, userAgent)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 429 {
		return resp, nil
	}
	resp.Body.Close()
	time.Sleep(2 * time.Second)
	cs.Throttle()
	return doRequest(client, url, userAgent)
}

func doRequest(client *http.Client, url, userAgent string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 && resp.StatusCode != 429 {
		resp.Body.Close()
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return resp, nil
}
