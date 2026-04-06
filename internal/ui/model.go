package ui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	"github.com/curkan/billmind/internal/domain"
	"github.com/curkan/billmind/internal/i18n"
	"github.com/curkan/billmind/internal/platform"
	"github.com/curkan/billmind/internal/storage"
)

// ViewMode represents the current screen of the application.
type ViewMode int

const (
	ViewList ViewMode = iota
	ViewWizard
	ViewEdit
	ViewDelete
	ViewConfirmPaid
	ViewFilter
	ViewHelp
	ViewSettings
)

// Model is the top-level Bubbletea model holding all application state.
type Model struct {
	viewMode     ViewMode
	table        *TableWidget
	helpViewport viewport.Model
	storage      *storage.Storage
	platform     platform.Platform
	reminders    []domain.Reminder

	// Search
	searchInput  textinput.Model
	searchActive bool
	searchOpen   bool

	// Tag filter
	activeFilters  []string
	filterCursor   int
	allTags        []string
	pendingFilters []string

	// Wizard (add reminder)
	wizard *WizardState

	// Edit overlay
	edit *EditState

	// Settings
	settingsCursor   int
	settingsSection  int // 0 = language, 1 = ntfy topic
	ntfyTopicInput   textinput.Model

	// Pending keys (vim-style combos)
	gPending bool
	dPending bool

	// State
	err        error
	width      int
	height     int
	statusMsg  string
	lastAction *UndoAction
}

// UndoType identifies the kind of action that can be undone.
type UndoType int

const (
	UndoDelete UndoType = iota
	UndoPaid
	UndoEdit
)

// UndoAction stores enough information to reverse the last destructive action.
type UndoAction struct {
	Type     UndoType
	Reminder domain.Reminder
}

// getTableColumns returns column definitions for the reminder table.
func getTableColumns() []TableColumn {
	return []TableColumn{
		{Title: "", Width: 2},                          // status symbol
		{Title: i18n.T("wizard.name"), Width: 22},      // name
		{Title: i18n.T("list.header_due"), Width: 16},  // due date
		{Title: i18n.T("wizard.interval"), Width: 14},  // interval
		{Title: i18n.T("wizard.tags"), Width: 16},      // tags
		{Title: "", Width: 2},                          // link icon
	}
}

// New creates a new Model wired to the given storage and platform.
func New(store *storage.Storage, plat platform.Platform) Model {
	tw := NewTableWidget()
	tw.SetColumns(getTableColumns())
	tw.SetRows([]TableRow{})

	si := textinput.New()
	si.Prompt = "/ "
	si.CharLimit = 100

	ntfyIn := textinput.New()
	ntfyIn.Placeholder = "billmind-myname123"
	ntfyIn.CharLimit = 100
	ntfyIn.SetWidth(30)

	return Model{
		viewMode:       ViewList,
		table:          tw,
		storage:        store,
		platform:       plat,
		searchInput:    si,
		ntfyTopicInput: ntfyIn,
	}
}

// Init returns the initial command that loads reminders and settings from disk.
func (m Model) Init() tea.Cmd {
	return tea.Batch(m.loadReminders, m.loadSettings)
}

func (m Model) loadReminders() tea.Msg {
	return loadRemindersMsg{}
}

func (m Model) loadSettings() tea.Msg {
	return settingsLoadMsg{}
}

type loadRemindersMsg struct{}

type settingsLoadMsg struct{}

type settingsLoadedMsg struct {
	settings storage.Settings
}
