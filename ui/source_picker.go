package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/JonathanWThom/feedme/api"
)

// Source picker options
var sourceOptions = []string{"Hacker News", "Lobste.rs", "Reddit"}

// handleSourcePickerInput handles keyboard input in the source picker
func (m Model) handleSourcePickerInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.editingSubreddit {
		switch msg.Type {
		case tea.KeyEnter:
			if m.subredditInput != "" {
				m.source = api.NewRedditClient(m.subredditInput)
				m.resetForNewSource()
				m.editingSubreddit = false
				return m, tea.Batch(m.spinner.Tick, m.loadStoryIDs())
			}
		case tea.KeyEsc:
			m.editingSubreddit = false
			m.subredditInput = ""
		case tea.KeyBackspace:
			if len(m.subredditInput) > 0 {
				m.subredditInput = m.subredditInput[:len(m.subredditInput)-1]
			}
		case tea.KeyRunes:
			m.subredditInput += string(msg.Runes)
		}
		return m, nil
	}

	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, m.keys.Up):
		if m.sourcePickerCursor > 0 {
			m.sourcePickerCursor--
		}

	case key.Matches(msg, m.keys.Down):
		if m.sourcePickerCursor < len(sourceOptions)-1 {
			m.sourcePickerCursor++
		}

	case key.Matches(msg, m.keys.Enter):
		switch m.sourcePickerCursor {
		case 0: // Hacker News
			m.source = api.NewClient()
			m.resetForNewSource()
			return m, tea.Batch(m.spinner.Tick, m.loadStoryIDs())
		case 1: // Lobste.rs
			m.source = api.NewLobstersClient()
			m.resetForNewSource()
			return m, tea.Batch(m.spinner.Tick, m.loadStoryIDs())
		case 2: // Reddit
			m.editingSubreddit = true
			m.subredditInput = ""
		}

	case key.Matches(msg, m.keys.Back):
		m.view = StoriesView
	}

	return m, nil
}

// renderSourcePicker renders the source selection view
func (m Model) renderSourcePicker() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(HeaderStyle.Render(" Switch Source "))
	b.WriteString("\n\n")

	for i, option := range sourceOptions {
		cursor := "  "
		if i == m.sourcePickerCursor {
			cursor = "> "
		}

		if i == m.sourcePickerCursor {
			b.WriteString(SelectedTitleStyle.Render(cursor + option))
		} else {
			b.WriteString(TitleStyle.Render(cursor + option))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")

	if m.editingSubreddit {
		b.WriteString(MetaStyle.Render("  Enter subreddit: r/"))
		b.WriteString(SelectedTitleStyle.Render(m.subredditInput))
		b.WriteString(SelectedTitleStyle.Render("_"))
		b.WriteString("\n\n")
		b.WriteString(MetaStyle.Render("  Press Enter to confirm, Esc to cancel"))
	} else {
		b.WriteString(MetaStyle.Render("  ↑↓: navigate  Enter: select  Esc: cancel"))
	}

	b.WriteString("\n")

	return b.String()
}
