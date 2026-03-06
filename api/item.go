package api

import (
	"fmt"
	"time"
)

// Item represents a news item (story, comment, job, poll)
type Item struct {
	ID          int    `json:"id"`
	Type        string `json:"type"`
	By          string `json:"by"`
	Time        int64  `json:"time"`
	Text        string `json:"text"`
	URL         string `json:"url"`
	Title       string `json:"title"`
	Score       int    `json:"score"`
	Kids        []int  `json:"kids"`
	Parent      int    `json:"parent"`
	Descendants int    `json:"descendants"`
	Deleted     bool   `json:"deleted"`
	Dead        bool   `json:"dead"`
}

// TimeAgo returns a human-readable time ago string
func (i *Item) TimeAgo() string {
	d := time.Since(time.Unix(i.Time, 0))
	hours := d.Hours()

	switch {
	case hours >= 24*365:
		return pluralizeAgo(int(hours/(24*365)), "year")
	case hours >= 24*30:
		return pluralizeAgo(int(hours/(24*30)), "month")
	case hours >= 24:
		return pluralizeAgo(int(hours/24), "day")
	case hours >= 1:
		return pluralizeAgo(int(hours), "hour")
	case d.Minutes() >= 1:
		return pluralizeAgo(int(d.Minutes()), "minute")
	default:
		return "just now"
	}
}

func pluralizeAgo(n int, unit string) string {
	if n == 1 {
		return fmt.Sprintf("1 %s ago", unit)
	}
	return fmt.Sprintf("%d %ss ago", n, unit)
}

// Domain extracts the domain from the URL
func (i *Item) Domain() string {
	if i.URL == "" {
		return ""
	}
	return extractHost(i.URL)
}

func extractHost(url string) string {
	url = stripProtocol(url)
	if len(url) > 4 && url[:4] == "www." {
		url = url[4:]
	}
	for i, c := range url {
		if c == '/' {
			return url[:i]
		}
	}
	return url
}

func stripProtocol(url string) string {
	if len(url) > 8 && url[:8] == "https://" {
		return url[8:]
	}
	if len(url) > 7 && url[:7] == "http://" {
		return url[7:]
	}
	return url
}

// Comment represents a comment with its depth for rendering
type Comment struct {
	*Item
	Depth    int
	Children []*Comment
}
