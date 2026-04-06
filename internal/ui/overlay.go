package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// overlayCenter places the foreground string centered on top of the background string,
// replacing the corresponding lines in the background.
func overlayCenter(bg, fg string, width, height int) string {
	fgHeight := len(strings.Split(fg, "\n"))
	fgWidth := lipgloss.Width(fg)

	startRow := (height - fgHeight) / 2
	startCol := (width - fgWidth) / 2

	return overlayAt(bg, fg, width, height, startRow, startCol)
}

// overlayBottomRight places the foreground string in the bottom-right corner
// on top of the background string, with a small margin.
func overlayBottomRight(bg, fg string, width, height int) string {
	fgHeight := len(strings.Split(fg, "\n"))
	fgWidth := lipgloss.Width(fg)

	margin := 1
	startRow := height - fgHeight - margin
	startCol := width - fgWidth - margin

	return overlayAt(bg, fg, width, height, startRow, startCol)
}

// overlayAt places the foreground string at a specific row/col position
// on top of the background string, replacing the corresponding lines.
func overlayAt(bg, fg string, width, height, startRow, startCol int) string {
	bgLines := strings.Split(bg, "\n")
	fgLines := strings.Split(fg, "\n")

	// Pad background to fill the terminal height
	for len(bgLines) < height {
		bgLines = append(bgLines, "")
	}

	if startRow < 0 {
		startRow = 0
	}
	if startCol < 0 {
		startCol = 0
	}

	for i, fgLine := range fgLines {
		row := startRow + i
		if row >= len(bgLines) {
			break
		}

		bgLine := bgLines[row]
		bgRunes := []rune(bgLine)

		// Pad background line if shorter than startCol
		for len(bgRunes) < startCol+lipgloss.Width(fgLine) {
			bgRunes = append(bgRunes, ' ')
		}

		// Build the composited line: bg prefix + fg overlay + bg suffix
		prefix := string(bgRunes[:startCol])
		fgLineWidth := lipgloss.Width(fgLine)
		suffixStart := startCol + fgLineWidth
		suffix := ""
		if suffixStart < len(bgRunes) {
			suffix = string(bgRunes[suffixStart:])
		}

		bgLines[row] = prefix + fgLine + suffix
	}

	return strings.Join(bgLines[:height], "\n")
}
