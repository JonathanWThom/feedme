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
	flag.StringVar(&sourceFlag, "source", "hn", "News source to use: hn (Hacker News) or lobsters (Lobste.rs)")
	flag.StringVar(&sourceFlag, "s", "hn", "News source to use (shorthand)")
	flag.Parse()

	var source api.Source
	switch strings.ToLower(sourceFlag) {
	case "hn", "hackernews", "hacker-news":
		source = api.NewClient()
	case "lobsters", "lobste.rs", "l":
		source = api.NewLobstersClient()
	default:
		fmt.Fprintf(os.Stderr, "Unknown source: %s\n", sourceFlag)
		fmt.Fprintf(os.Stderr, "Valid sources: hn, lobsters\n")
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
