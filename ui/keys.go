package ui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines all keybindings
type KeyMap struct {
	Up           key.Binding
	Down         key.Binding
	Enter        key.Binding
	Back         key.Binding
	Comments     key.Binding
	Open         key.Binding
	NextTab      key.Binding
	PrevTab      key.Binding
	Refresh      key.Binding
	Help         key.Binding
	Quit         key.Binding
	PageDown     key.Binding
	PageUp       key.Binding
	Home         key.Binding
	End          key.Binding
	ToggleMouse  key.Binding
	SwitchSource key.Binding
	Visual       key.Binding
	Yank         key.Binding
}

// DefaultKeyMap returns the default keybindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "open link"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc", "b"),
			key.WithHelp("esc/b", "back"),
		),
		Comments: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "comments"),
		),
		Open: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "open in browser"),
		),
		NextTab: key.NewBinding(
			key.WithKeys("tab", "l"),
			key.WithHelp("tab/l", "next feed"),
		),
		PrevTab: key.NewBinding(
			key.WithKeys("shift+tab", "h"),
			key.WithHelp("shift+tab/h", "prev feed"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown", "ctrl+d"),
			key.WithHelp("pgdn", "page down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup", "ctrl+u"),
			key.WithHelp("pgup", "page up"),
		),
		Home: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("home/g", "go to top"),
		),
		End: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("end/G", "go to bottom"),
		),
		ToggleMouse: key.NewBinding(
			key.WithKeys("m"),
			key.WithHelp("m", "toggle mouse (for copy)"),
		),
		SwitchSource: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "switch source"),
		),
		Visual: key.NewBinding(
			key.WithKeys("v"),
			key.WithHelp("v", "visual mode"),
		),
		Yank: key.NewBinding(
			key.WithKeys("y"),
			key.WithHelp("y", "yank selection"),
		),
	}
}

// ShortHelp returns a short help string
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Enter, k.Comments, k.NextTab, k.Help, k.Quit}
}

// FullHelp returns the full help strings
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.PageUp, k.PageDown, k.Home, k.End},
		{k.Enter, k.Open, k.Comments, k.Back},
		{k.NextTab, k.PrevTab, k.Refresh, k.SwitchSource},
		{k.Visual, k.Yank, k.ToggleMouse, k.Help, k.Quit},
	}
}
