package ui

import "charm.land/bubbles/v2/key"

type keyMap struct {
	Up      key.Binding
	Down    key.Binding
	Top     key.Binding
	Bottom  key.Binding
	Add     key.Binding
	Edit    key.Binding
	Delete  key.Binding
	Paid    key.Binding
	OpenURL key.Binding
	Undo    key.Binding
	Search  key.Binding
	Filter  key.Binding
	Help    key.Binding
	Quit    key.Binding
}

var keys = keyMap{
	Up:      key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("k", "up")),
	Down:    key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("j", "down")),
	Top:     key.NewBinding(key.WithKeys("g"), key.WithHelp("gg", "top")),
	Bottom:  key.NewBinding(key.WithKeys("G"), key.WithHelp("G", "bottom")),
	Add:     key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "add")),
	Edit:    key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit")),
	Delete:  key.NewBinding(key.WithKeys("d"), key.WithHelp("dd", "delete")),
	Paid:    key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "paid")),
	OpenURL: key.NewBinding(key.WithKeys("o"), key.WithHelp("o", "open url")),
	Undo:    key.NewBinding(key.WithKeys("u"), key.WithHelp("u", "undo")),
	Search:  key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search")),
	Filter:  key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "filter")),
	Help:    key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	Quit:    key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
}

// ShortHelp returns the key bindings shown in the compact help bar.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Add, k.Edit, k.Delete, k.Paid, k.Search, k.Filter, k.Help, k.Quit}
}

// FullHelp returns grouped key bindings for the expanded help view.
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Top, k.Bottom},
		{k.Add, k.Edit, k.Delete, k.Paid, k.OpenURL, k.Undo},
		{k.Search, k.Filter, k.Help, k.Quit},
	}
}
