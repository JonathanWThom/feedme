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

// version is set at build time via ldflags
var version = "dev"

func main() {
	var sourceFlag string
	var showVersion bool
	flag.StringVar(&sourceFlag, "source", "hn", "News source: hn, lobsters, r/subreddit, or RSS URL")
	flag.StringVar(&sourceFlag, "s", "hn", "News source (shorthand)")
	flag.BoolVar(&showVersion, "version", false, "Show version information")
	flag.BoolVar(&showVersion, "v", false, "Show version information (shorthand)")
	flag.Parse()

	if showVersion {
		fmt.Printf("feedme %s\n", version)
		os.Exit(0)
	}

	// Check for updates in background
	updateChan := make(chan *api.UpdateInfo, 1)
	go func() {
		updateChan <- api.CheckForUpdate(version)
	}()

	var source api.Source
	sourceLower := strings.ToLower(sourceFlag)

	switch {
	case sourceLower == "hn" || sourceLower == "hackernews" || sourceLower == "hacker-news":
		source = api.NewClient()
	case sourceLower == "lobsters" || sourceLower == "lobste.rs" || sourceLower == "l":
		source = api.NewLobstersClient()
	case strings.HasPrefix(sourceLower, "r/") || strings.HasPrefix(sourceLower, "/r/"):
		source = api.NewRedditClient(sourceFlag)
	case strings.HasPrefix(sourceLower, "http://") || strings.HasPrefix(sourceLower, "https://"):
		source = api.NewRSSClient(sourceFlag)
	default:
		fmt.Fprintf(os.Stderr, "Unknown source: %s\n", sourceFlag)
		fmt.Fprintf(os.Stderr, "Valid sources: hn, lobsters, r/subreddit, or RSS feed URL\n")
		os.Exit(1)
	}

	p := tea.NewProgram(
		ui.NewWithSource(source, updateChan),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
