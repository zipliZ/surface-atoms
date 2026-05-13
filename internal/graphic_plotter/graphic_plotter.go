package graphic_plotter

import (
	"fmt"
	"io"
	"log/slog"
	"main/configs"
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
	GraphicsToPlot []configs.GraphicToPlot
}

func New(exelFilePath string, outputFilePath string, lineLabel string, graphicsToPlot []configs.GraphicToPlot) *GraphicPlotter {
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
		// check if this header matches any xAxis exactly
		for _, graphicAxis := range p.GraphicsToPlot {
			if header == graphicAxis.XAxis {
				columnsIndexes[i] = header
				break
			}
		}
		// check if this header matches any yAxis (exact or suffix)
		for _, graphicAxis := range p.GraphicsToPlot {
			if header == graphicAxis.YAxis || (len(header) > len(graphicAxis.YAxis) && header[len(header)-len(graphicAxis.YAxis):] == graphicAxis.YAxis) {
				columnsIndexes[i] = header
				break
			}
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
		charts.WithTooltipOpts(opts.Tooltip{Show: opts.Bool(true), Trigger: "axis"}),
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(true)}),
	)

	line.SetXAxis(columnValues[xName])

	for colName, values := range columnValues {
		if colName == yName || (len(colName) > len(yName) && colName[len(colName)-len(yName):] == yName) {
			seriesName := colName
			line.AddSeries(seriesName, generateLineItems(values)).
				SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{Smooth: opts.Bool(true)}))
		}
	}

	return line, nil
}
