package tui

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"results_tui/internal/results"
)

func (m Model) previewChart(width, height int) (string, error) {
	if m.xAxis == "" || m.yAxis == "" {
		return "", errors.New("select X and Y columns to preview")
	}

	files := m.selectedFiles()
	if len(files) == 0 {
		return "", errors.New("select at least one valid file")
	}

	file := files[0]
	columns, err := results.ReadExcelColumns(file.Path, []string{m.xAxis, m.yAxis}, m.mergeSuffix)
	if err != nil {
		return "", err
	}
	x := results.FirstColumn(columns, m.xAxis, m.mergeSuffix)
	y := results.FirstColumn(columns, m.yAxis, m.mergeSuffix)
	if len(x) == 0 || len(y) == 0 {
		return "", errors.New("current columns are not available in the first selected file")
	}

	return asciiPlot(x, y, width, height), nil
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
