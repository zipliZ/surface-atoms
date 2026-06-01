package exporter

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"

	"results_tui/internal/domain"
	"results_tui/internal/results"
)

type seriesData struct {
	Name   string
	Points []opts.LineData
}

func ExportHTML(root string, files []domain.ResultFile, graphs []domain.PlotConfig, output string) (string, error) {
	if len(files) == 0 {
		return "", errors.New("no selected files")
	}
	if len(graphs) == 0 {
		return "", errors.New("no selected graphs")
	}
	if filepath.Ext(output) == "" {
		output += ".html"
	}
	if !filepath.IsAbs(output) {
		output = filepath.Join(root, output)
	}

	page := components.NewPage()
	page.PageTitle = "Combined Surface Atoms Results"

	var chartsToAdd []components.Charter
	var skipped []string
	for _, graph := range graphs {
		line, count := buildLineChart(files, graph)
		if count == 0 {
			skipped = append(skipped, graph.Label()+": no series")
			continue
		}
		chartsToAdd = append(chartsToAdd, line)
	}
	if len(chartsToAdd) == 0 {
		return "", fmt.Errorf("no charts generated: %s", strings.Join(skipped, "; "))
	}

	page.AddCharts(chartsToAdd...)
	file, err := os.Create(output)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if err := page.Render(io.MultiWriter(file)); err != nil {
		return "", err
	}

	abs, err := filepath.Abs(output)
	if err != nil {
		return output, nil
	}
	return abs, nil
}

func buildLineChart(files []domain.ResultFile, graph domain.PlotConfig) (*charts.Line, int) {
	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Theme: types.ThemeInfographic, PageTitle: "Combined Results", Width: "1200px", Height: "600px"}),
		charts.WithTitleOpts(opts.Title{Title: fmt.Sprintf("%s / %s", graph.YAxis, graph.XAxis)}),
		charts.WithYAxisOpts(opts.YAxis{Name: graph.YAxis, Scale: opts.Bool(true), SplitNumber: 10}),
		charts.WithXAxisOpts(opts.XAxis{Name: graph.XAxis, Type: "value", SplitNumber: 15}),
		charts.WithDataZoomOpts(opts.DataZoom{Type: "inside", Start: 0, End: 100, XAxisIndex: []int{0}}),
		charts.WithTooltipOpts(opts.Tooltip{Show: opts.Bool(true), Trigger: "axis"}),
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(true), Top: "30px", Type: "scroll"}),
	)

	seriesCount := 0
	for _, file := range files {
		columns, err := results.ReadExcelColumns(file.Path, []string{graph.XAxis, graph.YAxis}, graph.MergeSuffix)
		if err != nil {
			continue
		}
		xValues := results.FirstColumn(columns, graph.XAxis, graph.MergeSuffix)
		if len(xValues) == 0 {
			continue
		}
		for _, series := range resolveSeries(file, columns, xValues, graph.YAxis, graph.MergeSuffix) {
			line.AddSeries(series.Name, series.Points).
				SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{Smooth: opts.Bool(true), ShowSymbol: opts.Bool(false)}))
			seriesCount++
		}
	}
	return line, seriesCount
}

func resolveSeries(file domain.ResultFile, columns map[string][]float64, xValues []float64, yAxis string, mergeSuffix bool) []seriesData {
	var names []string
	if values, ok := columns[yAxis]; ok && len(values) > 0 {
		names = append(names, yAxis)
	}
	if mergeSuffix {
		for name, values := range columns {
			if name != yAxis && domain.NormalizeSuffix(name) == yAxis && len(values) > 0 {
				names = append(names, name)
			}
		}
	}
	sort.Strings(names)

	series := make([]seriesData, 0, len(names))
	for _, name := range names {
		label := formatSeriesLabel(file, name, yAxis, len(names) > 1 || name != yAxis)
		series = append(series, seriesData{Name: label, Points: generateXYItems(xValues, columns[name])})
	}
	return series
}

func formatSeriesLabel(file domain.ResultFile, concreteColumn, logicalColumn string, includeColumn bool) string {
	parts := make([]string, 0, 2)
	if file.Temperature != "" {
		parts = append(parts, file.Temperature)
	}
	if file.RunLabel != "" {
		parts = append(parts, file.RunLabel)
	}
	if len(parts) == 0 {
		parts = append(parts, strings.TrimSuffix(file.FileName, filepath.Ext(file.FileName)))
	}

	label := "[" + strings.Join(parts, " ") + "]"
	if includeColumn {
		return label + " " + concreteColumn
	}
	return label + " " + logicalColumn
}

func generateXYItems(xValues, yValues []float64) []opts.LineData {
	minLen := min(len(xValues), len(yValues))
	items := make([]opts.LineData, 0, minLen)
	for i := 0; i < minLen; i++ {
		items = append(items, opts.LineData{Value: []interface{}{xValues[i], yValues[i]}})
	}
	return items
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
