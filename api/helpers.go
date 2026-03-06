package api

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// hashShortID creates a pseudo-numeric ID from a short ID string
func hashShortID(shortID string) int {
	hash := 0
	for i, c := range shortID {
		hash += int(c) * (i + 1) * 31
	}
	return hash
}

// parseTime attempts to parse a timestamp string in common formats
func parseTime(s string) (time.Time, error) {
	formats := []string{
		"2006-01-02 15:04:05 -0700",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05-07:00",
		time.RFC3339,
	}
	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unable to parse time: %s", s)
}

// parseRelativeTime converts relative time strings like "3 hours ago"
// to a Unix timestamp
func parseRelativeTime(s string) int64 {
	now := time.Now()
	s = strings.ToLower(strings.TrimSpace(s))

	re := regexp.MustCompile(`(\d+)\s*(second|minute|hour|day|week|month|year)s?\s*ago`)
	matches := re.FindStringSubmatch(s)
	if len(matches) < 3 {
		return now.Unix()
	}

	num, err := strconv.Atoi(matches[1])
	if err != nil {
		return now.Unix()
	}

	var duration time.Duration
	switch matches[2] {
	case "second":
		duration = time.Duration(num) * time.Second
	case "minute":
		duration = time.Duration(num) * time.Minute
	case "hour":
		duration = time.Duration(num) * time.Hour
	case "day":
		duration = time.Duration(num) * 24 * time.Hour
	case "week":
		duration = time.Duration(num) * 7 * 24 * time.Hour
	case "month":
		duration = time.Duration(num) * 30 * 24 * time.Hour
	case "year":
		duration = time.Duration(num) * 365 * 24 * time.Hour
	}

	return now.Add(-duration).Unix()
}
