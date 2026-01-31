package api

// NPR RSS feeds
// NPR provides RSS feeds for news and culture

var nprFeeds = map[string]string{
	"news":     "https://feeds.npr.org/1001/rss.xml", // News
	"world":    "https://feeds.npr.org/1004/rss.xml", // World
	"us":       "https://feeds.npr.org/1003/rss.xml", // National/US
	"politics": "https://feeds.npr.org/1014/rss.xml", // Politics
	"business": "https://feeds.npr.org/1006/rss.xml", // Business
	"tech":     "https://feeds.npr.org/1019/rss.xml", // Technology
	"science":  "https://feeds.npr.org/1007/rss.xml", // Science
	"health":   "https://feeds.npr.org/1128/rss.xml", // Health
	"culture":  "https://feeds.npr.org/1008/rss.xml", // Arts & Life
}

var nprFeedNames = []string{"news", "world", "us", "politics", "business", "tech", "science", "health", "culture"}
var nprFeedLabels = []string{"News", "World", "US", "Politics", "Business", "Tech", "Science", "Health", "Culture"}

// NPRClient is a client for NPR RSS feeds
type NPRClient struct {
	*RSSClient
}

// NewNPRClient creates a new NPR client
func NewNPRClient() *NPRClient {
	return &NPRClient{
		RSSClient: NewRSSClient("NPR", "https://www.npr.org", nprFeeds, nprFeedNames, nprFeedLabels),
	}
}

// StoryURL returns the NPR article URL
func (c *NPRClient) StoryURL(item *Item) string {
	return item.URL
}
