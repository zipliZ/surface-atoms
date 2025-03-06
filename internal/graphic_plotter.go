package internal

import (
	"fmt"
	"io"
	"log/slog"
	"main/internal/config"
	"os"
	"strconv"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
	"github.com/xuri/excelize/v2"
)

type GraphicPlotter struct {
	ExelFilePath   string
	OutputFilePath string
	LineLabel      string
	GraphicsToPlot []config.GraphicToPlot
}

func NewGraphicPlotter(exelFilePath string, outputFilePath string, lineLabel string, graphicsToPlot []config.GraphicToPlot) *GraphicPlotter {
	return &GraphicPlotter{
		ExelFilePath:   exelFilePath,
		OutputFilePath: outputFilePath,
		GraphicsToPlot: graphicsToPlot,
		LineLabel:      lineLabel,
	}
}

func (p *GraphicPlotter) Plot() error {
	if len(p.GraphicsToPlot) == 0 {
		return nil
	}
	slog.Info("start plot graphics", "count", len(p.GraphicsToPlot))
	columnValues, err := p.readExcel()
	if err != nil {
		return err
	}
	linesGraphics := make([]components.Charter, 0)
	for _, graphicAxis := range p.GraphicsToPlot {
		line, plotErr := p.plotGraph(columnValues, graphicAxis.XAxis, graphicAxis.YAxis)
		if plotErr != nil {
			slog.Error("plot error", "err", plotErr)
		}
		linesGraphics = append(linesGraphics, line)
	}

	page := components.NewPage()
	page.PageTitle = "Surface Atoms"
	page.AddCharts(
		linesGraphics...,
	)
	f, err := os.Create(p.OutputFilePath)
	if err != nil {
		return err
	}

	return page.Render(io.MultiWriter(f))
}

func (p *GraphicPlotter) readExcel() (columnValues map[string][]float64, err error) {
	file, err := excelize.OpenFile(p.ExelFilePath)
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

	requiredColumns := map[string]struct{}{}
	for _, graphicAxis := range p.GraphicsToPlot {
		if _, ok := requiredColumns[graphicAxis.XAxis]; !ok {
			requiredColumns[graphicAxis.XAxis] = struct{}{}
		}
		if _, ok := requiredColumns[graphicAxis.YAxis]; !ok {
			requiredColumns[graphicAxis.YAxis] = struct{}{}
		}
	}

	columnsIndexes := make(map[int]string)
	for i, header := range rows[0] {
		if _, ok := requiredColumns[header]; ok {
			columnsIndexes[i] = header
		}
	}

	columnValues = make(map[string][]float64)
	// Читаем данные по индексам колонок
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		for index, columnName := range columnsIndexes {
			value, _ := strconv.ParseFloat(row[index], 64)
			columnValues[columnName] = append(columnValues[columnName], value)
		}
	}

	return
}

func generateLineItems(yValues []float64) []opts.LineData {
	items := make([]opts.LineData, 0)
	for _, value := range yValues {
		items = append(items, opts.LineData{Value: value})
	}
	return items
}

func (p *GraphicPlotter) plotGraph(columnValues map[string][]float64, xName, yName string) (*charts.Line, error) {
	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Theme: types.ThemeEssos, PageTitle: "Surface Atoms", Width: "1200px"}),
		charts.WithTitleOpts(opts.Title{
			Title: fmt.Sprintf("%s/%s", yName, xName),
		}), charts.WithYAxisOpts(opts.YAxis{
			Name:        yName,
			Scale:       opts.Bool(true),
			SplitNumber: 10,
		}),
		charts.WithXAxisOpts(opts.XAxis{
			Name:        xName,
			SplitNumber: 15,
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:       "inside",
			Start:      0,
			End:        100,
			XAxisIndex: []int{0},
		}),
	)

	// Put data into instance
	line.SetXAxis(columnValues[xName]).
		AddSeries(p.LineLabel, generateLineItems(columnValues[yName])).
		SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{Smooth: opts.Bool(true)}))

	return line, nil
}
