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
		return m.handleSubredditInput(msg)
	}
	return m.handleSourcePickerNav(msg)
}

func (m Model) handleSubredditInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		return m.confirmSubreddit()
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

func (m Model) confirmSubreddit() (tea.Model, tea.Cmd) {
	if m.subredditInput == "" {
		return m, nil
	}
	m.source = api.NewRedditClient(m.subredditInput)
	m.resetForNewSource()
	m.editingSubreddit = false
	return m, tea.Batch(m.spinner.Tick, m.loadStoryIDs())
}

func (m Model) handleSourcePickerNav(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
		return m.selectSource()
	case key.Matches(msg, m.keys.Back):
		m.view = StoriesView
	}
	return m, nil
}

func (m Model) selectSource() (tea.Model, tea.Cmd) {
	switch m.sourcePickerCursor {
	case 0:
		m.source = api.NewClient()
	case 1:
		m.source = api.NewLobstersClient()
	case 2:
		m.editingSubreddit = true
		m.subredditInput = ""
		return m, nil
	}
	m.resetForNewSource()
	return m, tea.Batch(m.spinner.Tick, m.loadStoryIDs())
}

// renderSourcePicker renders the source selection view
func (m Model) renderSourcePicker() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(HeaderStyle.Render(" Switch Source "))
	b.WriteString("\n\n")
	b.WriteString(m.renderSourceOptions())
	b.WriteString("\n")
	b.WriteString(m.renderSourcePickerFooter())
	b.WriteString("\n")
	return b.String()
}

func (m Model) renderSourceOptions() string {
	var b strings.Builder
	for i, option := range sourceOptions {
		selected := i == m.sourcePickerCursor
		cursor := "  "
		if selected {
			cursor = "> "
		}
		if selected {
			b.WriteString(SelectedTitleStyle.Render(cursor + option))
		} else {
			b.WriteString(TitleStyle.Render(cursor + option))
		}
		b.WriteString("\n")
	}
	return b.String()
}

func (m Model) renderSourcePickerFooter() string {
	if !m.editingSubreddit {
		return MetaStyle.Render("  ↑↓: navigate  Enter: select  Esc: cancel")
	}
	return MetaStyle.Render("  Enter subreddit: r/") +
		SelectedTitleStyle.Render(m.subredditInput) +
		SelectedTitleStyle.Render("_") +
		"\n\n" +
		MetaStyle.Render("  Press Enter to confirm, Esc to cancel")
}
