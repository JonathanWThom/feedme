package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	orange     = lipgloss.Color("#FF6600")
	dimOrange  = lipgloss.Color("#CC5500")
	subtle     = lipgloss.Color("#666666")
	highlight  = lipgloss.Color("#FFFFFF")
	dimText    = lipgloss.Color("#888888")
	commentBg  = lipgloss.Color("#1a1a1a")

	// Header
	HeaderStyle = lipgloss.NewStyle().
			Background(orange).
			Foreground(lipgloss.Color("#000000")).
			Bold(true).
			Padding(0, 1)

	TabStyle = lipgloss.NewStyle().
			Foreground(dimText).
			Padding(0, 1)

	ActiveTabStyle = lipgloss.NewStyle().
			Foreground(highlight).
			Bold(true).
			Padding(0, 1).
			Underline(true)

	// Story list
	TitleStyle = lipgloss.NewStyle().
			Foreground(highlight).
			Bold(true)

	SelectedTitleStyle = lipgloss.NewStyle().
				Foreground(orange).
				Bold(true)

	URLStyle = lipgloss.NewStyle().
			Foreground(subtle)

	MetaStyle = lipgloss.NewStyle().
			Foreground(dimText)

	ScoreStyle = lipgloss.NewStyle().
			Foreground(orange)

	// Comments
	CommentAuthorStyle = lipgloss.NewStyle().
				Foreground(orange).
				Bold(true)

	CommentTextStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#CCCCCC"))

	CommentMetaStyle = lipgloss.NewStyle().
				Foreground(dimText)

	// General
	HelpStyle = lipgloss.NewStyle().
			Foreground(dimText)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000"))

	SpinnerStyle = lipgloss.NewStyle().
			Foreground(orange)

	// Status bar
	StatusBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#333333")).
			Foreground(dimText).
			Padding(0, 1)
)

// IndentStyle returns a style for comment indentation
func IndentStyle(depth int) lipgloss.Style {
	colors := []lipgloss.Color{
		orange,
		lipgloss.Color("#4A9EFF"),
		lipgloss.Color("#50C878"),
		lipgloss.Color("#FFD700"),
		lipgloss.Color("#FF69B4"),
		lipgloss.Color("#9370DB"),
	}
	color := colors[depth%len(colors)]
	return lipgloss.NewStyle().
		Foreground(color).
		Bold(true)
}
