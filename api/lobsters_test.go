package api

import (
	"testing"
	"time"
)

func TestHashShortID(t *testing.T) {
	tests := []struct {
		name string
		id   string
	}{
		{"short id", "abc123"},
		{"another id", "xyz789"},
		{"single char", "a"},
		{"empty string", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hashShortID(tt.id)
			// Deterministic: same input produces same output
			got2 := hashShortID(tt.id)
			if got != got2 {
				t.Errorf("hashShortID(%q) not deterministic: %d != %d", tt.id, got, got2)
			}
		})
	}

	// Different inputs should (generally) produce different outputs
	a := hashShortID("abc")
	b := hashShortID("xyz")
	if a == b {
		t.Errorf("hashShortID produced same value for different inputs: %d", a)
	}
}

func TestParseTime(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"offset format", "2024-01-15 10:30:00 -0500", false},
		{"UTC format", "2024-01-15T10:30:00Z", false},
		{"timezone offset", "2024-01-15T10:30:00-05:00", false},
		{"RFC3339", "2024-01-15T10:30:00Z", false},
		{"invalid format", "not a date", true},
		{"empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseTime(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseTime(%q) expected error, got %v", tt.input, result)
				}
			} else {
				if err != nil {
					t.Errorf("parseTime(%q) unexpected error: %v", tt.input, err)
				}
				if result.IsZero() {
					t.Errorf("parseTime(%q) returned zero time", tt.input)
				}
			}
		})
	}
}

func TestParseRelativeTime(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		input     string
		wantDelta time.Duration
		tolerance time.Duration
	}{
		{"5 minutes ago", "5 minutes ago", 5 * time.Minute, 2 * time.Second},
		{"1 hour ago", "1 hour ago", 1 * time.Hour, 2 * time.Second},
		{"3 hours ago", "3 hours ago", 3 * time.Hour, 2 * time.Second},
		{"2 days ago", "2 days ago", 2 * 24 * time.Hour, 2 * time.Second},
		{"1 week ago", "1 week ago", 7 * 24 * time.Hour, 2 * time.Second},
		{"1 month ago", "1 month ago", 30 * 24 * time.Hour, 2 * time.Second},
		{"1 year ago", "1 year ago", 365 * 24 * time.Hour, 2 * time.Second},
		{"30 seconds ago", "30 seconds ago", 30 * time.Second, 2 * time.Second},
		{"invalid returns now", "invalid text", 0, 2 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseRelativeTime(tt.input)
			expected := now.Add(-tt.wantDelta).Unix()
			diff := got - expected
			if diff < 0 {
				diff = -diff
			}
			toleranceSecs := int64(tt.tolerance.Seconds())
			if diff > toleranceSecs {
				t.Errorf("parseRelativeTime(%q) = %d, want ~%d (diff: %ds)",
					tt.input, got, expected, diff)
			}
		})
	}
}
