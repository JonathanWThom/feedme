package api

import (
	"testing"
	"time"
)

func TestItem_TimeAgo(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name string
		time int64
		want string
	}{
		{"just now", now.Unix(), "just now"},
		{"30 seconds ago", now.Add(-30 * time.Second).Unix(), "just now"},
		{"1 minute ago", now.Add(-1 * time.Minute).Unix(), "1 minute ago"},
		{"5 minutes ago", now.Add(-5 * time.Minute).Unix(), "5 minutes ago"},
		{"1 hour ago", now.Add(-1 * time.Hour).Unix(), "1 hour ago"},
		{"3 hours ago", now.Add(-3 * time.Hour).Unix(), "3 hours ago"},
		{"1 day ago", now.Add(-24 * time.Hour).Unix(), "1 day ago"},
		{"7 days ago", now.Add(-7 * 24 * time.Hour).Unix(), "7 days ago"},
		{"1 month ago", now.Add(-31 * 24 * time.Hour).Unix(), "1 month ago"},
		{"3 months ago", now.Add(-90 * 24 * time.Hour).Unix(), "3 months ago"},
		{"1 year ago", now.Add(-366 * 24 * time.Hour).Unix(), "1 year ago"},
		{"2 years ago", now.Add(-730 * 24 * time.Hour).Unix(), "2 years ago"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := &Item{Time: tt.time}
			got := item.TimeAgo()
			if got != tt.want {
				t.Errorf("TimeAgo() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestItem_Domain(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{"empty url", "", ""},
		{"https with path", "https://example.com/foo/bar", "example.com"},
		{"http with path", "http://example.com/foo", "example.com"},
		{"https with www", "https://www.example.com/foo", "example.com"},
		{"bare domain no slash", "https://example.com", "example.com"},
		{"subdomain preserved", "https://blog.example.com/post", "blog.example.com"},
		{"www subdomain stripped", "https://www.blog.example.com/x", "blog.example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := &Item{URL: tt.url}
			got := item.Domain()
			if got != tt.want {
				t.Errorf("Domain() = %q, want %q", got, tt.want)
			}
		})
	}
}
