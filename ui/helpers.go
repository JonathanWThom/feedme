package ui

import (
	"html"
	"regexp"
	"strings"
)

// stripAnsi removes ANSI escape codes from a string
func stripAnsi(s string) string {
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return ansiRegex.ReplaceAllString(s, "")
}

func cleanHTML(s string) string {
	s = html.UnescapeString(s)
	s = regexp.MustCompile(`<p>`).ReplaceAllString(s, "\n\n")
	s = regexp.MustCompile(`<br\s*/?\s*>`).ReplaceAllString(s, "\n")
	s = regexp.MustCompile(`<a\s+href="([^"]*)"[^>]*>([^<]*)</a>`).ReplaceAllString(s, "$2 [$1]")
	s = regexp.MustCompile(`<[^>]*>`).ReplaceAllString(s, "")
	s = strings.TrimSpace(s)
	return s
}

func wrapText(s string, width int) string {
	return strings.Join(wrapTextLines(s, width), "\n")
}

func wrapTextLines(s string, width int) []string {
	if width <= 0 {
		width = 80
	}

	var lines []string
	paragraphs := strings.Split(s, "\n")

	for _, para := range paragraphs {
		if para == "" {
			lines = append(lines, "")
			continue
		}

		words := strings.Fields(para)
		if len(words) == 0 {
			lines = append(lines, "")
			continue
		}

		currentLine := words[0]
		for _, word := range words[1:] {
			if len(currentLine)+1+len(word) <= width {
				currentLine += " " + word
			} else {
				lines = append(lines, currentLine)
				currentLine = word
			}
		}
		lines = append(lines, currentLine)
	}

	return lines
}
