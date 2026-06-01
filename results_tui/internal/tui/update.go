package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"results_tui/internal/domain"
	"results_tui/internal/exporter"
	"results_tui/internal/openfile"
)

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+left":
			m.moveFocusLeft()
			return m, nil
		case "ctrl+right":
			m.moveFocusRight()
			return m, nil
		case "ctrl+up":
			m.moveFocusUp()
			return m, nil
		case "ctrl+down":
			m.moveFocusDown()
			return m, nil
		}

		if m.focus == panelOutput {
			switch msg.String() {
			case "esc", "enter":
				m.output.Blur()
				m.setFocus(panelFiles)
				return m, nil
			case "tab", "shift+tab", "ctrl+c":
			default:
				var cmd tea.Cmd
				m.output, cmd = m.output.Update(msg)
				return m, cmd
			}
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab":
			m.setFocus((m.focus + 1) % 4)
			return m, nil
		case "shift+tab":
			m.setFocus((m.focus + 3) % 4)
			return m, nil
		case "up", "k":
			m.moveCursor(-1)
			return m, nil
		case "down", "j":
			m.moveCursor(1)
			return m, nil
		case " ":
			m.toggleSelection()
			return m, nil
		case "x":
			m.setColumnAxis(true)
			return m, nil
		case "y":
			m.setColumnAxis(false)
			return m, nil
		case "m":
			m.mergeSuffix = !m.mergeSuffix
			m.refreshColumns()
			return m, nil
		case "a":
			m.addGraph()
			return m, nil
		case "d", "backspace", "delete":
			m.deleteGraph()
			return m, nil
		case "e":
			return m.export()
		case "o":
			if m.lastExport == "" {
				m.status = "Export a HTML file first."
				return m, nil
			}
			m.status = "Opening exported HTML..."
			return m, openCmd(m.lastExport)
		}
	case exportMsg:
		if msg.err != nil {
			m.status = "Export failed: " + msg.err.Error()
			return m, nil
		}
		m.lastExport = msg.path
		m.status = "Exported " + msg.path + ". Press o to open."
		return m, nil
	case openMsg:
		if msg.err != nil {
			m.status = "Open failed: " + msg.err.Error()
		} else {
			m.status = "Open command started."
		}
		return m, nil
	}
	return m, nil
}

func (m *Model) setFocus(panelID panel) {
	m.focus = panelID
	m.syncOutputFocus()
}

func (m *Model) syncOutputFocus() {
	if m.focus == panelOutput {
		m.output.Focus()
		return
	}
	m.output.Blur()
}

func (m *Model) moveFocusLeft() {
	switch m.focus {
	case panelColumns:
		m.setFocus(panelFiles)
	case panelGraphs:
		m.setFocus(panelColumns)
	case panelOutput:
		m.setFocus(panelFiles)
	default:
		m.setFocus(panelGraphs)
	}
}

func (m *Model) moveFocusRight() {
	switch m.focus {
	case panelFiles:
		m.setFocus(panelColumns)
	case panelColumns:
		m.setFocus(panelGraphs)
	case panelOutput:
		m.setFocus(panelGraphs)
	default:
		m.setFocus(panelFiles)
	}
}

func (m *Model) moveFocusUp() {
	if m.focus == panelOutput {
		m.setFocus(panelColumns)
	}
}

func (m *Model) moveFocusDown() {
	if m.focus != panelOutput {
		m.setFocus(panelOutput)
	}
}

func (m *Model) moveCursor(delta int) {
	switch m.focus {
	case panelFiles:
		m.fileCursor = clampCursor(m.fileCursor+delta, len(m.files))
	case panelColumns:
		m.colCursor = clampCursor(m.colCursor+delta, len(m.columns))
		if len(m.columns) > 0 {
			m.yAxis = m.columns[m.colCursor].Name
		}
	case panelGraphs:
		m.graphCursor = clampCursor(m.graphCursor+delta, len(m.graphs))
	}
}

func (m *Model) toggleSelection() {
	switch m.focus {
	case panelFiles:
		if len(m.files) == 0 {
			return
		}
		file := m.files[m.fileCursor]
		if !file.Valid() {
			m.status = "Cannot select invalid file: " + file.ReadError.Error()
			return
		}
		if m.selected[m.fileCursor] {
			delete(m.selected, m.fileCursor)
		} else {
			m.selected[m.fileCursor] = true
		}
		m.refreshColumns()
	case panelColumns:
		if len(m.columns) == 0 {
			return
		}
		m.setColumnAxis(false)
	}
}

func (m *Model) setColumnAxis(asX bool) {
	if m.focus != panelColumns || len(m.columns) == 0 {
		return
	}
	name := m.columns[m.colCursor].Name
	if asX {
		m.xAxis = name
		if m.yAxis == name {
			m.yAxis = ""
			for _, col := range m.columns {
				if col.Name != m.xAxis {
					m.yAxis = col.Name
					break
				}
			}
		}
		return
	}
	if name == m.xAxis {
		m.status = "Y column must be different from X."
		return
	}
	m.yAxis = name
}

func (m *Model) addGraph() {
	if m.xAxis == "" || m.yAxis == "" {
		m.status = "Select both X and Y columns before adding a graph."
		return
	}
	if m.xAxis == m.yAxis {
		m.status = "X and Y columns must be different."
		return
	}

	graph := domain.PlotConfig{XAxis: m.xAxis, YAxis: m.yAxis, MergeSuffix: m.mergeSuffix}
	for _, existing := range m.graphs {
		if existing == graph {
			m.status = "This graph is already selected."
			return
		}
	}
	m.graphs = append(m.graphs, graph)
	m.graphCursor = len(m.graphs) - 1
	m.status = "Graph added: " + graph.Label()
}

func (m *Model) deleteGraph() {
	if m.focus != panelGraphs || len(m.graphs) == 0 {
		return
	}
	m.graphs = append(m.graphs[:m.graphCursor], m.graphs[m.graphCursor+1:]...)
	m.graphCursor = clampCursor(m.graphCursor, len(m.graphs))
	m.status = "Graph removed."
}

func (m Model) export() (tea.Model, tea.Cmd) {
	if len(m.graphs) == 0 {
		m.status = "Add at least one graph before export."
		return m, nil
	}
	selected := m.selectedFiles()
	if len(selected) == 0 {
		m.status = "Select at least one valid result file before export."
		return m, nil
	}

	output := strings.TrimSpace(m.output.Value())
	if output == "" {
		output = domain.DefaultOutputFile
	}
	m.status = "Exporting HTML..."
	return m, exportCmd(m.root, selected, m.graphs, output)
}

func exportCmd(root string, files []domain.ResultFile, graphs []domain.PlotConfig, output string) tea.Cmd {
	return func() tea.Msg {
		path, err := exporter.ExportHTML(root, files, graphs, output)
		return exportMsg{path: path, err: err}
	}
}

func openCmd(path string) tea.Cmd {
	return func() tea.Msg {
		return openMsg{err: openfile.Open(path)}
	}
}

func clampCursor(value, size int) int {
	if size <= 0 {
		return 0
	}
	if value < 0 {
		return size - 1
	}
	if value >= size {
		return 0
	}
	return value
}
