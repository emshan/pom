package main

import (
	"github.com/charmbracelet/bubbles/key"
)

type KeyMap struct {
	Quit       key.Binding
	SwitchView key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		SwitchView: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "switch view"),
		),
	}
}

func (k KeyMap) ShortHelp() string {
	return "tab: switch view • q: quit"
}
