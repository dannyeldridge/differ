package main

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Up         key.Binding
	Down       key.Binding
	FocusNext  key.Binding
	FocusPrev  key.Binding
	GotoTop    key.Binding
	GotoBottom key.Binding
	Quit       key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("k", "up"),
		key.WithHelp("k/↑", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("j", "down"),
		key.WithHelp("j/↓", "down"),
	),
	FocusNext: key.NewBinding(
		key.WithKeys("l", "right", "tab"),
		key.WithHelp("l/→/tab", "focus next pane"),
	),
	FocusPrev: key.NewBinding(
		key.WithKeys("h", "left", "shift+tab"),
		key.WithHelp("h/←/shift+tab", "focus prev pane"),
	),
	GotoTop: key.NewBinding(
		key.WithKeys("g"),
		key.WithHelp("g", "go to top"),
	),
	GotoBottom: key.NewBinding(
		key.WithKeys("G"),
		key.WithHelp("G", "go to bottom"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}
