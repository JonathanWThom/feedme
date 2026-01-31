package api

// Google News RSS feeds
// Google News provides topic-based RSS feeds

var googleNewsFeeds = map[string]string{
	"top":       "https://news.google.com/rss?hl=en-US&gl=US&ceid=US:en",
	"world":     "https://news.google.com/rss/topics/CAAqJggKIiBDQkFTRWdvSUwyMHZNRGx1YlY4U0FtVnVHZ0pWVXlnQVAB?hl=en-US&gl=US&ceid=US:en",
	"us":        "https://news.google.com/rss/topics/CAAqIggKIhxDQkFTRHdvSkwyMHZNRGxqTjNjd0VnSmxiaWdBUAE?hl=en-US&gl=US&ceid=US:en",
	"business":  "https://news.google.com/rss/topics/CAAqJggKIiBDQkFTRWdvSUwyMHZNRGx6TVdZU0FtVnVHZ0pWVXlnQVAB?hl=en-US&gl=US&ceid=US:en",
	"tech":      "https://news.google.com/rss/topics/CAAqJggKIiBDQkFTRWdvSUwyMHZNRGRqTVhZU0FtVnVHZ0pWVXlnQVAB?hl=en-US&gl=US&ceid=US:en",
	"science":   "https://news.google.com/rss/topics/CAAqJggKIiBDQkFTRWdvSUwyMHZNRFp0Y1RjU0FtVnVHZ0pWVXlnQVAB?hl=en-US&gl=US&ceid=US:en",
	"health":    "https://news.google.com/rss/topics/CAAqIQgKIhtDQkFTRGdvSUwyMHZNR3QwTlRFU0FtVnVLQUFQAQ?hl=en-US&gl=US&ceid=US:en",
	"sports":    "https://news.google.com/rss/topics/CAAqJggKIiBDQkFTRWdvSUwyMHZNRFp1ZEdvU0FtVnVHZ0pWVXlnQVAB?hl=en-US&gl=US&ceid=US:en",
	"entertain": "https://news.google.com/rss/topics/CAAqJggKIiBDQkFTRWdvSUwyMHZNREpxYW5RU0FtVnVHZ0pWVXlnQVAB?hl=en-US&gl=US&ceid=US:en",
}

var googleNewsFeedNames = []string{"top", "world", "us", "business", "tech", "science", "health", "sports", "entertain"}
var googleNewsFeedLabels = []string{"Top", "World", "US", "Business", "Tech", "Science", "Health", "Sports", "Arts"}

// GoogleNewsClient is a client for Google News RSS feeds
type GoogleNewsClient struct {
	*RSSClient
}

// NewGoogleNewsClient creates a new Google News client
func NewGoogleNewsClient() *GoogleNewsClient {
	return &GoogleNewsClient{
		RSSClient: NewRSSClient("Google News", "https://news.google.com", googleNewsFeeds, googleNewsFeedNames, googleNewsFeedLabels),
	}
}

// StoryURL returns the article URL
func (c *GoogleNewsClient) StoryURL(item *Item) string {
	return item.URL
}
