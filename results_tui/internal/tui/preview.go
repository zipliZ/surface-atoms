package tui

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"results_tui/internal/domain"
	"results_tui/internal/results"
)

func (m Model) previewChart(width, height int) (string, error) {
	graph, err := m.previewGraph()
	if err != nil {
		return "", err
	}

	files := m.selectedFiles()
	if len(files) == 0 {
		return "", errors.New("select at least one valid file")
	}

	file := files[0]
	columns, err := results.ReadExcelColumns(file.Path, []string{graph.XAxis, graph.YAxis}, graph.MergeSuffix)
	if err != nil {
		return "", err
	}
	x := results.FirstColumn(columns, graph.XAxis, graph.MergeSuffix)
	y := results.FirstColumn(columns, graph.YAxis, graph.MergeSuffix)
	if len(x) == 0 || len(y) == 0 {
		return "", errors.New("current columns are not available in the first selected file")
	}

	return asciiPlot(x, y, width, height), nil
}

func (m Model) previewGraph() (domain.PlotConfig, error) {
	if m.focus == panelGraphs && len(m.graphs) > 0 {
		cursor := clampCursor(m.graphCursor, len(m.graphs))
		return m.graphs[cursor], nil
	}

	if m.xAxis == "" || m.yAxis == "" {
		return domain.PlotConfig{}, errors.New("select X and Y columns to preview")
	}
	return domain.PlotConfig{XAxis: m.xAxis, YAxis: m.yAxis, MergeSuffix: m.mergeSuffix}, nil
}

func asciiPlot(xValues, yValues []float64, width, height int) string {
	width = max(10, width)
	height = max(4, height)
	minLen := min(len(xValues), len(yValues))
	if minLen == 0 {
		return ""
	}
	if minLen > width {
		step := float64(minLen-1) / float64(width-1)
		sampled := make([]float64, 0, width)
		for i := 0; i < width; i++ {
			sampled = append(sampled, yValues[int(math.Round(float64(i)*step))])
		}
		yValues = sampled
		minLen = len(sampled)
	}

	minY, maxY := yValues[0], yValues[0]
	for i := 0; i < minLen; i++ {
		minY = math.Min(minY, yValues[i])
		maxY = math.Max(maxY, yValues[i])
	}
	if minY == maxY {
		maxY = minY + 1
	}

	canvas := make([][]rune, height)
	for row := range canvas {
		canvas[row] = make([]rune, minLen)
		for col := range canvas[row] {
			canvas[row][col] = ' '
		}
	}
	for col := 0; col < minLen; col++ {
		normalized := (yValues[col] - minY) / (maxY - minY)
		row := height - 1 - int(math.Round(normalized*float64(height-1)))
		row = min(max(row, 0), height-1)
		canvas[row][col] = '*'
	}

	lines := make([]string, 0, height+1)
	for _, row := range canvas {
		lines = append(lines, string(row))
	}
	lines = append(lines, fmt.Sprintf("y %.4g..%.4g", minY, maxY))
	return strings.Join(lines, "\n")
}
