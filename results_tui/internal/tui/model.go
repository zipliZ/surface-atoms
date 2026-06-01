package tui

import (
	"path/filepath"

	"github.com/charmbracelet/bubbles/textinput"

	"results_tui/internal/configdefaults"
	"results_tui/internal/domain"
)

type panel int

const (
	panelFiles panel = iota
	panelColumns
	panelGraphs
	panelOutput
)

type Model struct {
	root        string
	files       []domain.ResultFile
	selected    map[int]bool
	focus       panel
	fileCursor  int
	colCursor   int
	graphCursor int

	mergeSuffix bool
	xAxis       string
	yAxis       string
	columns     []domain.ColumnOption

	graphs []domain.PlotConfig
	output textinput.Model

	status       string
	lastExport   string
	width        int
	height       int
	configLoaded bool
}

type exportMsg struct {
	path string
	err  error
}

type openMsg struct {
	err error
}

func NewModel(root string, files []domain.ResultFile) Model {
	ti := textinput.New()
	ti.Placeholder = domain.DefaultOutputFile
	ti.SetValue(domain.DefaultOutputFile)
	ti.CharLimit = 240
	ti.Width = 40

	m := Model{
		root:        root,
		files:       files,
		selected:    make(map[int]bool),
		focus:       panelFiles,
		mergeSuffix: false,
		xAxis:       domain.DefaultXAxis,
		output:      ti,
		status:      "Select result files, choose columns, add graphs, then export.",
	}

	for i, file := range files {
		if file.Valid() {
			m.selected[i] = true
		}
	}
	m.refreshColumns()
	m.loadConfigDefaults()
	return m
}

func (m *Model) refreshColumns() {
	m.columns = domain.ComputeColumns(m.selectedFiles(), m.mergeSuffix)
	if len(m.columns) == 0 {
		m.xAxis = ""
		m.yAxis = ""
		m.colCursor = 0
		return
	}

	names := make(map[string]bool, len(m.columns))
	for _, col := range m.columns {
		names[col.Name] = true
	}
	if m.xAxis == "" || !names[m.xAxis] {
		m.xAxis = m.columns[0].Name
		if names[domain.DefaultXAxis] {
			m.xAxis = domain.DefaultXAxis
		}
	}
	if m.yAxis == "" || !names[m.yAxis] || m.yAxis == m.xAxis {
		m.yAxis = ""
		for _, col := range m.columns {
			if col.Name != m.xAxis {
				m.yAxis = col.Name
				break
			}
		}
	}
	for i, col := range m.columns {
		if col.Name == m.yAxis {
			m.colCursor = i
			break
		}
	}
}

func (m *Model) selectedFiles() []domain.ResultFile {
	var selected []domain.ResultFile
	for index, file := range m.files {
		if m.selected[index] && file.Valid() {
			selected = append(selected, file)
		}
	}
	return selected
}

func (m *Model) loadConfigDefaults() {
	strictColumns := domain.ComputeColumns(m.selectedFiles(), false)
	graphs, loaded := configdefaults.Load(filepath.Join(m.root, "config.yaml"), strictColumns)
	m.configLoaded = loaded
	if len(graphs) == 0 {
		return
	}

	m.graphs = append(m.graphs, graphs...)
	m.status = "Loaded default graphs from config.yaml."
}
