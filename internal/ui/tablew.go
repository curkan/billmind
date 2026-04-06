package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// TableColumn defines a single column in the table widget.
type TableColumn struct {
	Title string
	Width int
}

// TableRow holds the cell values and an optional foreground style for the entire row.
type TableRow struct {
	Cells []string
	Style lipgloss.Style
}

// TableWidget is a minimal table renderer with full control over per-row coloring.
// Unlike bubbles/v2/table, it composes each row as plain text first, then applies
// a single lipgloss.Render call so that ANSI reset codes from inner cells cannot
// break the selected-row background highlight.
type TableWidget struct {
	columns []TableColumn
	rows    []TableRow
	cursor  int
	height  int
	width   int
	offset  int

	headerStyle   lipgloss.Style
	selectedStyle lipgloss.Style
}

// NewTableWidget creates a TableWidget with sensible default styles.
func NewTableWidget() *TableWidget {
	return &TableWidget{
		headerStyle: lipgloss.NewStyle().Bold(true).Foreground(colorHint),
		selectedStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#F8F8F2")).
			Background(lipgloss.Color("#44475A")),
		height: 20,
	}
}

// SetColumns replaces the column definitions.
func (t *TableWidget) SetColumns(cols []TableColumn) { t.columns = cols }

// SetRows replaces all rows and clamps the cursor.
func (t *TableWidget) SetRows(rows []TableRow) { t.rows = rows; t.clampCursor() }

// SetHeight sets the number of visible rows (excluding the header).
func (t *TableWidget) SetHeight(h int) { t.height = h; t.updateOffset() }

// SetWidth sets the total rendering width used for full-width row highlighting.
func (t *TableWidget) SetWidth(w int) { t.width = w }

// Cursor returns the current cursor position.
func (t *TableWidget) Cursor() int { return t.cursor }

// SetCursor moves the cursor to position n, clamped to valid bounds.
func (t *TableWidget) SetCursor(n int) { t.cursor = n; t.clampCursor() }

// Rows returns the current row data.
func (t *TableWidget) Rows() []TableRow { return t.rows }

// MoveDown moves the cursor down by n rows.
func (t *TableWidget) MoveDown(n int) { t.cursor += n; t.clampCursor() }

// MoveUp moves the cursor up by n rows.
func (t *TableWidget) MoveUp(n int) { t.cursor -= n; t.clampCursor() }

// GotoTop moves the cursor to the first row.
func (t *TableWidget) GotoTop() { t.cursor = 0; t.updateOffset() }

// GotoBottom moves the cursor to the last row.
func (t *TableWidget) GotoBottom() {
	t.cursor = len(t.rows) - 1
	t.clampCursor()
}

func (t *TableWidget) clampCursor() {
	if t.cursor < 0 {
		t.cursor = 0
	}
	if t.cursor >= len(t.rows) {
		t.cursor = len(t.rows) - 1
	}
	if t.cursor < 0 {
		t.cursor = 0
	}
	t.updateOffset()
}

func (t *TableWidget) updateOffset() {
	if t.cursor < t.offset {
		t.offset = t.cursor
	}
	if t.cursor >= t.offset+t.height {
		t.offset = t.cursor - t.height + 1
	}
	if t.offset < 0 {
		t.offset = 0
	}
}

// View renders the table header and visible rows as a string.
func (t *TableWidget) View() string {
	var b strings.Builder

	b.WriteString(t.renderHeader())
	b.WriteString("\n")

	end := t.offset + t.height
	if end > len(t.rows) {
		end = len(t.rows)
	}

	for i := t.offset; i < end; i++ {
		b.WriteString(t.renderRow(i))
		if i < end-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

// renderHeader builds the header line from column titles using plain-text padding.
func (t *TableWidget) renderHeader() string {
	var line strings.Builder
	for _, col := range t.columns {
		line.WriteString(" ")
		title := ansi.Truncate(col.Title, col.Width, "")
		line.WriteString(title)
		pad := col.Width - ansi.StringWidth(title)
		if pad > 0 {
			line.WriteString(strings.Repeat(" ", pad))
		}
		line.WriteString(" ")
	}
	return t.headerStyle.Render(line.String())
}

// renderRow builds a single data row. The cells are composed as pure plain text
// (spaces for padding, no ANSI codes), then a SINGLE lipgloss.Render call wraps
// the entire line, applying foreground and/or background uniformly.
func (t *TableWidget) renderRow(idx int) string {
	row := t.rows[idx]
	isSelected := idx == t.cursor

	var line strings.Builder
	for i, col := range t.columns {
		value := ""
		if i < len(row.Cells) {
			value = row.Cells[i]
		}
		// Strip any existing ANSI codes to guarantee plain text
		value = ansi.Strip(value)
		// Truncate to column width
		truncated := ansi.Truncate(value, col.Width, "…")
		// Pad with spaces
		cellWidth := ansi.StringWidth(truncated)
		pad := col.Width - cellWidth
		if pad < 0 {
			pad = 0
		}
		line.WriteString(" ") // left padding
		line.WriteString(truncated)
		line.WriteString(strings.Repeat(" ", pad))
		line.WriteString(" ") // right padding
	}

	plainLine := line.String()

	if isSelected {
		style := t.selectedStyle
		if t.width > 0 {
			style = style.Width(t.width)
		}
		fg := row.Style.GetForeground()
		if fg != nil {
			style = style.Foreground(fg)
		}
		return style.Render(plainLine)
	}

	fg := row.Style.GetForeground()
	if fg != nil {
		style := row.Style
		if t.width > 0 {
			style = style.Width(t.width)
		}
		return style.Render(plainLine)
	}

	return plainLine
}
