package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	leftWidth := max(30, m.width/3-2)
	midWidth := max(34, m.width/3-2)
	rightWidth := max(30, m.width-leftWidth-midWidth-8)
	bottomHeight := min(18, max(12, m.height/3))
	bodyHeight := max(8, m.height-bottomHeight-5)

	top := lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.renderPanel(panelFiles, "Results", m.renderFiles(bodyHeight), leftWidth, bodyHeight),
		" ",
		m.renderPanel(panelColumns, "Columns", m.renderColumns(bodyHeight), midWidth, bodyHeight),
		" ",
		m.renderPanel(panelGraphs, "Graphs", m.renderGraphs(bodyHeight), rightWidth, bodyHeight),
	)

	outputWidth := max(20, m.width-4)
	bottom := m.renderPanel(panelOutput, "Preview and export", m.renderOutput(bottomHeight, outputWidth), outputWidth, bottomHeight)
	help := m.renderHelp()
	return lipgloss.JoinVertical(lipgloss.Left, top, bottom, help)
}

func (m Model) renderPanel(panelID panel, title, body string, width, height int) string {
	style := panelBorder
	if m.focus == panelID {
		style = focusedBorder
	}
	return style.Width(width).Height(height).Render(titleStyle.Render(strings.ToUpper(title)) + "\n\n" + body)
}

func (m Model) renderFiles(height int) string {
	if len(m.files) == 0 {
		return mutedStyle.Render("No .xlsx files found in result folders.")
	}

	lines := make([]string, 0, len(m.files)+2)
	for i, file := range m.files {
		cursor := " "
		if m.focus == panelFiles && i == m.fileCursor {
			cursor = ">"
		}
		check := "[ ]"
		if m.selected[i] {
			check = "[x]"
		}

		line := fmt.Sprintf("%s %s %s", cursor, check, file.Label())
		if !file.Valid() {
			line = fmt.Sprintf("%s [!] %s", cursor, file.Label())
			line = errorStyle.Render(line)
		} else if m.selected[i] {
			line = selectedStyle.Render(line)
		}
		lines = append(lines, line)
	}

	lines = append(lines, "", mutedStyle.Render(fmt.Sprintf("%d selected", len(m.selectedFiles()))))
	return fitLines(lines, height-3)
}

func (m Model) renderColumns(height int) string {
	mode := "strict"
	if m.mergeSuffix {
		mode = "suffix merge"
	}
	lines := []string{
		fmt.Sprintf("Mode: %s", mode),
		fmt.Sprintf("X: %s", emptyDash(m.xAxis)),
		fmt.Sprintf("Y: %s", emptyDash(m.yAxis)),
		"",
	}

	if len(m.columns) == 0 {
		lines = append(lines, mutedStyle.Render("No common numeric columns."))
		return fitLines(lines, height-3)
	}

	for i, col := range m.columns {
		cursor := " "
		if m.focus == panelColumns && i == m.colCursor {
			cursor = ">"
		}
		markers := "   "
		switch {
		case col.Name == m.xAxis && col.Name == m.yAxis:
			markers = "XY "
		case col.Name == m.xAxis:
			markers = "X  "
		case col.Name == m.yAxis:
			markers = "Y  "
		}

		line := fmt.Sprintf("%s %s%s", cursor, markers, col.Name)
		if m.focus == panelColumns && i == m.colCursor {
			line = cursorStyle.Render(line)
		}
		lines = append(lines, line)
	}
	return fitLines(lines, height-3)
}

func (m Model) renderGraphs(height int) string {
	if len(m.graphs) == 0 {
		return mutedStyle.Render("No graphs selected.\nPress a after choosing X/Y.")
	}

	lines := make([]string, 0, len(m.graphs))
	for i, graph := range m.graphs {
		cursor := " "
		if m.focus == panelGraphs && i == m.graphCursor {
			cursor = ">"
		}
		line := fmt.Sprintf("%s %d. %s", cursor, i+1, graph.Label())
		if m.focus == panelGraphs && i == m.graphCursor {
			line = cursorStyle.Render(line)
		}
		lines = append(lines, line)
	}
	return fitLines(lines, height-3)
}

func (m Model) renderOutput(height, width int) string {
	status := m.status
	if status == "" {
		status = "Ready."
	}
	lines := []string{
		"Output: " + m.output.View(),
		"Status: " + status,
	}
	if m.configLoaded {
		lines = append(lines, mutedStyle.Render("config.yaml defaults loaded when columns matched."))
	}

	errorsList := m.fileErrors()
	if len(errorsList) > 0 {
		lines = append(lines, "", errorStyle.Render("Skipped files:"))
		for _, item := range errorsList {
			lines = append(lines, errorStyle.Render(item))
			if len(lines) > 5 {
				break
			}
		}
	}

	lines = append(lines, "", "Preview:")
	maxBodyLines := max(1, height-3)
	plotHeight := max(3, maxBodyLines-len(lines)-1)
	preview, err := m.previewChart(min(120, max(20, width-6)), plotHeight)
	if err != nil {
		lines = append(lines, mutedStyle.Render(err.Error()))
	} else {
		lines = append(lines, preview)
	}
	return fitLines(lines, height-3)
}

func (m Model) fileErrors() []string {
	var items []string
	for _, file := range m.files {
		if file.ReadError != nil {
			items = append(items, fmt.Sprintf("%s: %v", file.FileName, file.ReadError))
		}
	}
	return items
}

func fitLines(lines []string, maxLines int) string {
	if maxLines <= 0 {
		return ""
	}
	if len(lines) > maxLines {
		lines = append(lines[:maxLines-1], mutedStyle.Render("..."))
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderHelp() string {
	lines := []string{
		"Navigation: Tab / Shift+Tab - next/previous block; Ctrl+Arrow - switch blocks; Up/Down or J/K - move inside block.",
		"Actions: Space - select file or set Y; X - set X column; M - strict/suffix mode; A - add graph; D - delete graph; E - export; O - open; Q - quit.",
	}
	return helpStyle.Render(strings.Join(lines, "\n"))
}

func emptyDash(value string) string {
	if value == "" {
		return "-"
	}
	return value
}
