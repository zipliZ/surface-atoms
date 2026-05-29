package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
	"github.com/xuri/excelize/v2"
	"gopkg.in/yaml.v3"
)

type fileConfig struct {
	Simulating simulatingConfig `yaml:"simulating"`
}

type simulatingConfig struct {
	GraphicsToPlot []graphicToPlot `yaml:"graphicsToPlot"`
}

type graphicToPlot struct {
	XAxis string `yaml:"xAxis"`
	YAxis string `yaml:"yAxis"`
}

func configLoad(path string) (fileConfig, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return fileConfig{}, err
	}

	var cfg fileConfig
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return fileConfig{}, err
	}

	return cfg, nil
}

type FileData struct {
	FileName     string
	ColumnValues map[string][]float64
}

func main() {
	cfg, err := configLoad("./../../config.yaml")
	if err != nil {
		// fallback if config can't be loaded, we can still use default graphics or exit
		log.Fatalf("failed to load configs: %v", err)
	}

	if len(cfg.Simulating.GraphicsToPlot) == 0 {
		log.Println("no graphics to plot in config")
		return
	}

	// Find all xlsx files in result directories
	var excelFiles []string
	entries, err := os.ReadDir(".")
	if err != nil {
		log.Fatalf("failed to read dir: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "result") {
			subEntries, err := os.ReadDir(entry.Name())
			if err != nil {
				continue
			}
			for _, sub := range subEntries {
				if !sub.IsDir() && strings.HasSuffix(sub.Name(), ".xlsx") {
					excelFiles = append(excelFiles, filepath.Join(entry.Name(), sub.Name()))
				}
			}
		}
	}

	if len(excelFiles) == 0 {
		log.Println("No excel files found in result folders.")
		return
	}

	var allData []FileData
	for _, file := range excelFiles {
		colValues, err := readExcel(file, cfg.Simulating.GraphicsToPlot)
		if err != nil {
			log.Printf("failed to read %s: %v", file, err)
			continue
		}
		allData = append(allData, FileData{
			FileName:     filepath.Base(file),
			ColumnValues: colValues,
		})
	}

	if len(allData) == 0 {
		log.Println("No data read from files.")
		return
	}

	linesGraphics := make([]components.Charter, 0)
	for _, graphicAxis := range cfg.Simulating.GraphicsToPlot {
		line := plotGraph(allData, graphicAxis.XAxis, graphicAxis.YAxis)
		linesGraphics = append(linesGraphics, line)
	}

	page := components.NewPage()
	page.PageTitle = "Combined Surface Atoms Results"
	page.AddCharts(linesGraphics...)

	outputFilePath := "combined_results.html"
	f, err := os.Create(outputFilePath)
	if err != nil {
		log.Fatalf("failed to create output file: %v", err)
	}
	defer f.Close()

	if err := page.Render(io.MultiWriter(f)); err != nil {
		log.Fatalf("failed to render combined html: %v", err)
	}

	log.Printf("Successfully generated combined combined_results.html with data from %d files", len(allData))
}

func readExcel(exelFilePath string, graphicsToPlot []graphicToPlot) (map[string][]float64, error) {
	file, err := excelize.OpenFile(exelFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	rows, err := file.GetRows("Sheet1")
	if err != nil {
		return nil, err
	}
	if len(rows) < 2 {
		return nil, fmt.Errorf("no data")
	}

	// keep track of required columns to avoid storing everything
	requiredColumns := map[string]struct{}{}
	for _, graphicAxis := range graphicsToPlot {
		requiredColumns[graphicAxis.XAxis] = struct{}{}
		requiredColumns[graphicAxis.YAxis] = struct{}{}
	}

	columnsIndexes := make(map[int]string)
	for i, header := range rows[0] {
		for _, graphicAxis := range graphicsToPlot {
			if header == graphicAxis.XAxis {
				columnsIndexes[i] = header
				break
			}
		}
		for _, graphicAxis := range graphicsToPlot {
			if header == graphicAxis.YAxis || (len(header) > len(graphicAxis.YAxis) && header[len(header)-len(graphicAxis.YAxis):] == graphicAxis.YAxis) {
				columnsIndexes[i] = header
				break
			}
		}
	}

	columnValues := make(map[string][]float64)
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		for index, columnName := range columnsIndexes {
			if index < len(row) {
				value, _ := strconv.ParseFloat(row[index], 64)
				columnValues[columnName] = append(columnValues[columnName], value)
			}
		}
	}

	return columnValues, nil
}

func generateXYItems(xValues, yValues []float64) []opts.LineData {
	items := make([]opts.LineData, 0)
	minLen := len(xValues)
	if len(yValues) < minLen {
		minLen = len(yValues)
	}
	for i := 0; i < minLen; i++ {
		items = append(items, opts.LineData{Value: []interface{}{xValues[i], yValues[i]}})
	}
	return items
}

func plotGraph(allData []FileData, xName, yName string) *charts.Line {
	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Theme: types.ThemeInfographic, PageTitle: "Combined Results", Width: "1200px", Height: "600px"}),
		charts.WithTitleOpts(opts.Title{
			Title: fmt.Sprintf("%s / %s", yName, xName),
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Name:        yName,
			Scale:       opts.Bool(true),
			SplitNumber: 10,
		}),
		charts.WithXAxisOpts(opts.XAxis{
			Name:        xName,
			Type:        "value",
			SplitNumber: 15,
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:       "inside",
			Start:      0,
			End:        100,
			XAxisIndex: []int{0},
		}),
		charts.WithTooltipOpts(opts.Tooltip{Show: opts.Bool(true), Trigger: "axis"}),
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(true), Top: "30px", Type: "scroll"}),
	)

	for _, fileData := range allData {
		xValues, hasX := fileData.ColumnValues[xName]
		if !hasX {
			continue
		}

		for colName, values := range fileData.ColumnValues {
			if colName == yName || (len(colName) > len(yName) && colName[len(colName)-len(yName):] == yName) {
				// We create a composite series name like "T300K - N - Qty"
				label := strings.ReplaceAll(fileData.FileName, ".xlsx", "")
				if strings.HasPrefix(label, "result_") {
					label = strings.TrimPrefix(label, "result_")
				}
				seriesName := fmt.Sprintf("[%s] %s", label, colName)

				line.AddSeries(seriesName, generateXYItems(xValues, values)).
					SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{Smooth: opts.Bool(true), ShowSymbol: opts.Bool(false)}))
			}
		}
	}

	return line
}
