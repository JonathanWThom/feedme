package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/JonathanWThom/feedme/api"
	"github.com/JonathanWThom/feedme/ui"
)

func main() {
	var sourceFlag string
	flag.StringVar(&sourceFlag, "source", "hn", "News source: hn, lobsters, or r/subreddit (e.g., r/golang)")
	flag.StringVar(&sourceFlag, "s", "hn", "News source (shorthand)")
	flag.Parse()

	var source api.Source
	sourceLower := strings.ToLower(sourceFlag)

	switch {
	case sourceLower == "hn" || sourceLower == "hackernews" || sourceLower == "hacker-news":
		source = api.NewClient()
	case sourceLower == "lobsters" || sourceLower == "lobste.rs" || sourceLower == "l":
		source = api.NewLobstersClient()
	case strings.HasPrefix(sourceLower, "r/") || strings.HasPrefix(sourceLower, "/r/"):
		source = api.NewRedditClient(sourceFlag)
	default:
		fmt.Fprintf(os.Stderr, "Unknown source: %s\n", sourceFlag)
		fmt.Fprintf(os.Stderr, "Valid sources: hn, lobsters, r/subreddit\n")
		os.Exit(1)
	}

	p := tea.NewProgram(
		ui.NewWithSource(source),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
