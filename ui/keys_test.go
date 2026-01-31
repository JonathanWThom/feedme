package ui

import (
	"testing"

	"github.com/charmbracelet/bubbles/key"
)

func TestDefaultKeyMap(t *testing.T) {
	km := DefaultKeyMap()

	// Test that all keybindings are properly initialized
	keybindings := []struct {
		name    string
		binding key.Binding
	}{
		{"Up", km.Up},
		{"Down", km.Down},
		{"Enter", km.Enter},
		{"Back", km.Back},
		{"Comments", km.Comments},
		{"Open", km.Open},
		{"NextTab", km.NextTab},
		{"PrevTab", km.PrevTab},
		{"Refresh", km.Refresh},
		{"Help", km.Help},
		{"Quit", km.Quit},
		{"PageDown", km.PageDown},
		{"PageUp", km.PageUp},
		{"Home", km.Home},
		{"End", km.End},
		{"ToggleMouse", km.ToggleMouse},
		{"SwitchSource", km.SwitchSource},
		{"Visual", km.Visual},
		{"Yank", km.Yank},
	}

	for _, kb := range keybindings {
		t.Run(kb.name, func(t *testing.T) {
			if len(kb.binding.Keys()) == 0 {
				t.Errorf("%s binding has no keys", kb.name)
			}
			if kb.binding.Help().Key == "" {
				t.Errorf("%s binding has no help key", kb.name)
			}
		})
	}
}

func TestUpKey(t *testing.T) {
	km := DefaultKeyMap()
	keys := km.Up.Keys()

	if !contains(keys, "up") {
		t.Error("Up binding should include 'up' key")
	}
	if !contains(keys, "k") {
		t.Error("Up binding should include 'k' key (vim style)")
	}
}

func TestDownKey(t *testing.T) {
	km := DefaultKeyMap()
	keys := km.Down.Keys()

	if !contains(keys, "down") {
		t.Error("Down binding should include 'down' key")
	}
	if !contains(keys, "j") {
		t.Error("Down binding should include 'j' key (vim style)")
	}
}

func TestQuitKey(t *testing.T) {
	km := DefaultKeyMap()
	keys := km.Quit.Keys()

	if !contains(keys, "q") {
		t.Error("Quit binding should include 'q' key")
	}
	if !contains(keys, "ctrl+c") {
		t.Error("Quit binding should include 'ctrl+c' key")
	}
}

func TestBackKey(t *testing.T) {
	km := DefaultKeyMap()
	keys := km.Back.Keys()

	if !contains(keys, "esc") {
		t.Error("Back binding should include 'esc' key")
	}
	if !contains(keys, "b") {
		t.Error("Back binding should include 'b' key")
	}
}

func TestTabKeys(t *testing.T) {
	km := DefaultKeyMap()

	nextKeys := km.NextTab.Keys()
	if !contains(nextKeys, "tab") {
		t.Error("NextTab binding should include 'tab' key")
	}
	if !contains(nextKeys, "l") {
		t.Error("NextTab binding should include 'l' key (vim style)")
	}

	prevKeys := km.PrevTab.Keys()
	if !contains(prevKeys, "shift+tab") {
		t.Error("PrevTab binding should include 'shift+tab' key")
	}
	if !contains(prevKeys, "h") {
		t.Error("PrevTab binding should include 'h' key (vim style)")
	}
}

func TestPageKeys(t *testing.T) {
	km := DefaultKeyMap()

	pgDownKeys := km.PageDown.Keys()
	if !contains(pgDownKeys, "pgdown") {
		t.Error("PageDown binding should include 'pgdown' key")
	}
	if !contains(pgDownKeys, "ctrl+d") {
		t.Error("PageDown binding should include 'ctrl+d' key (vim style)")
	}

	pgUpKeys := km.PageUp.Keys()
	if !contains(pgUpKeys, "pgup") {
		t.Error("PageUp binding should include 'pgup' key")
	}
	if !contains(pgUpKeys, "ctrl+u") {
		t.Error("PageUp binding should include 'ctrl+u' key (vim style)")
	}
}

func TestHomeEndKeys(t *testing.T) {
	km := DefaultKeyMap()

	homeKeys := km.Home.Keys()
	if !contains(homeKeys, "home") {
		t.Error("Home binding should include 'home' key")
	}
	if !contains(homeKeys, "g") {
		t.Error("Home binding should include 'g' key (vim style)")
	}

	endKeys := km.End.Keys()
	if !contains(endKeys, "end") {
		t.Error("End binding should include 'end' key")
	}
	if !contains(endKeys, "G") {
		t.Error("End binding should include 'G' key (vim style)")
	}
}

func TestActionKeys(t *testing.T) {
	km := DefaultKeyMap()

	// Enter
	enterKeys := km.Enter.Keys()
	if !contains(enterKeys, "enter") {
		t.Error("Enter binding should include 'enter' key")
	}

	// Comments
	commentsKeys := km.Comments.Keys()
	if !contains(commentsKeys, "c") {
		t.Error("Comments binding should include 'c' key")
	}

	// Open
	openKeys := km.Open.Keys()
	if !contains(openKeys, "o") {
		t.Error("Open binding should include 'o' key")
	}

	// Refresh
	refreshKeys := km.Refresh.Keys()
	if !contains(refreshKeys, "r") {
		t.Error("Refresh binding should include 'r' key")
	}

	// Help
	helpKeys := km.Help.Keys()
	if !contains(helpKeys, "?") {
		t.Error("Help binding should include '?' key")
	}

	// Mouse toggle
	mouseKeys := km.ToggleMouse.Keys()
	if !contains(mouseKeys, "m") {
		t.Error("ToggleMouse binding should include 'm' key")
	}

	// Switch source
	sourceKeys := km.SwitchSource.Keys()
	if !contains(sourceKeys, "s") {
		t.Error("SwitchSource binding should include 's' key")
	}

	// Visual mode
	visualKeys := km.Visual.Keys()
	if !contains(visualKeys, "v") {
		t.Error("Visual binding should include 'v' key")
	}

	// Yank
	yankKeys := km.Yank.Keys()
	if !contains(yankKeys, "y") {
		t.Error("Yank binding should include 'y' key")
	}
}

func TestShortHelp(t *testing.T) {
	km := DefaultKeyMap()
	shortHelp := km.ShortHelp()

	if len(shortHelp) == 0 {
		t.Error("ShortHelp should return some bindings")
	}

	// Should include essential bindings
	expectedCount := 7 // Up, Down, Enter, Comments, NextTab, Help, Quit
	if len(shortHelp) != expectedCount {
		t.Errorf("expected %d bindings in short help, got %d", expectedCount, len(shortHelp))
	}
}

func TestFullHelp(t *testing.T) {
	km := DefaultKeyMap()
	fullHelp := km.FullHelp()

	if len(fullHelp) == 0 {
		t.Error("FullHelp should return some binding groups")
	}

	// Should have 4 groups
	expectedGroups := 4
	if len(fullHelp) != expectedGroups {
		t.Errorf("expected %d groups in full help, got %d", expectedGroups, len(fullHelp))
	}

	// First group: navigation
	if len(fullHelp[0]) != 6 {
		t.Errorf("navigation group should have 6 bindings, got %d", len(fullHelp[0]))
	}

	// Second group: actions
	if len(fullHelp[1]) != 4 {
		t.Errorf("actions group should have 4 bindings, got %d", len(fullHelp[1]))
	}

	// Third group: feeds/source
	if len(fullHelp[2]) != 4 {
		t.Errorf("feeds group should have 4 bindings, got %d", len(fullHelp[2]))
	}

	// Fourth group: modes/quit
	if len(fullHelp[3]) != 5 {
		t.Errorf("modes group should have 5 bindings, got %d", len(fullHelp[3]))
	}
}

func TestHelpText(t *testing.T) {
	km := DefaultKeyMap()

	testCases := []struct {
		binding      key.Binding
		expectedKey  string
		expectedDesc string
	}{
		{km.Up, "↑/k", "up"},
		{km.Down, "↓/j", "down"},
		{km.Enter, "enter", "open link"},
		{km.Comments, "c", "comments"},
		{km.Help, "?", "help"},
		{km.Quit, "q", "quit"},
		{km.Visual, "v", "visual mode"},
		{km.Yank, "y", "yank selection"},
	}

	for _, tc := range testCases {
		t.Run(tc.expectedDesc, func(t *testing.T) {
			help := tc.binding.Help()
			if help.Key != tc.expectedKey {
				t.Errorf("expected key '%s', got '%s'", tc.expectedKey, help.Key)
			}
			if help.Desc != tc.expectedDesc {
				t.Errorf("expected desc '%s', got '%s'", tc.expectedDesc, help.Desc)
			}
		})
	}
}

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
