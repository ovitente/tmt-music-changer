package main

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Up         key.Binding
	Down       key.Binding
	Tab        key.Binding
	Enter      key.Binding
	Space      key.Binding
	Save       key.Binding
	Toggle     key.Binding
	Order      key.Binding
	Edit       key.Binding
	AddTrack   key.Binding
	Escape     key.Binding
	MoveUp     key.Binding
	MoveDown   key.Binding
	Delete     key.Binding
	Search     key.Binding
	SortToggle key.Binding
	Restore    key.Binding
	FileSwitch key.Binding
	Quit       key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter", "l"),
	),
	Space: key.NewBinding(
		key.WithKeys(" "),
	),
	Save: key.NewBinding(
		key.WithKeys("s"),
	),
	Toggle: key.NewBinding(
		key.WithKeys("t"),
	),
	Order: key.NewBinding(
		key.WithKeys("o"),
	),
	Edit: key.NewBinding(
		key.WithKeys("e"),
	),
	AddTrack: key.NewBinding(
		key.WithKeys("a"),
	),
	Escape: key.NewBinding(
		key.WithKeys("esc"),
	),
	MoveUp: key.NewBinding(
		key.WithKeys("K", "shift+up"),
	),
	MoveDown: key.NewBinding(
		key.WithKeys("J", "shift+down"),
	),
	Delete: key.NewBinding(
		key.WithKeys("d", "x"),
	),
	Search: key.NewBinding(
		key.WithKeys("f"),
	),
	SortToggle: key.NewBinding(
		key.WithKeys("r"),
	),
	Restore: key.NewBinding(
		key.WithKeys("R"),
	),
	FileSwitch: key.NewBinding(
		key.WithKeys("p"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
	),
}
