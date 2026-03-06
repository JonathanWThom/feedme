package ui

import "testing"

func TestCleanHTML(t *testing.T) {
	tests := []struct {
		name string
		html string
		want string
	}{
		{"plain text", "hello world", "hello world"},
		{"paragraph tags", "first<p>second", "first\n\nsecond"},
		{"br tags", "line1<br>line2", "line1\nline2"},
		{"self-closing br", "line1<br/>line2", "line1\nline2"},
		{"br with space", "line1<br />line2", "line1\nline2"},
		{
			"link extraction",
			`click <a href="https://example.com">here</a> now`,
			"click here [https://example.com] now",
		},
		{"html entities", "foo &amp; bar &lt; baz", "foo & bar < baz"},
		{"strip unknown tags", "hello <b>bold</b> world", "hello bold world"},
		{"leading/trailing whitespace", "  hello  ", "hello"},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanHTML(tt.html)
			if got != tt.want {
				t.Errorf("cleanHTML(%q) = %q, want %q", tt.html, got, tt.want)
			}
		})
	}
}

func TestWrapTextLines(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		width int
		want  []string
	}{
		{
			"short line no wrap",
			"hello world",
			80,
			[]string{"hello world"},
		},
		{
			"wraps at width",
			"one two three four five",
			10,
			[]string{"one two", "three four", "five"},
		},
		{
			"preserves paragraph breaks",
			"first\n\nsecond",
			80,
			[]string{"first", "", "second"},
		},
		{
			"zero width defaults to 80",
			"hello",
			0,
			[]string{"hello"},
		},
		{
			"negative width defaults to 80",
			"hello",
			-1,
			[]string{"hello"},
		},
		{
			"single long word",
			"superlongword",
			5,
			[]string{"superlongword"},
		},
		{
			"empty string",
			"",
			80,
			[]string{""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wrapTextLines(tt.text, tt.width)
			if len(got) != len(tt.want) {
				t.Errorf("wrapTextLines() returned %d lines, want %d\ngot:  %q\nwant: %q",
					len(got), len(tt.want), got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("line %d: got %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestStripAnsi(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"no ansi", "hello", "hello"},
		{"bold", "\x1b[1mhello\x1b[0m", "hello"},
		{"color", "\x1b[31mred\x1b[0m text", "red text"},
		{"multiple codes", "\x1b[1;31mbold red\x1b[0m", "bold red"},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripAnsi(tt.in)
			if got != tt.want {
				t.Errorf("stripAnsi(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
