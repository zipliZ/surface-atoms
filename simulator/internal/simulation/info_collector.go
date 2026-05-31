package simulation

import (
	"fmt"
	"main/configs"
	"math"
	"os"

	"github.com/tealeg/xlsx"
)

// InfoCollector - structure that collects information about the simulation progress.
type InfoCollector struct {
	fileName        string
	floatPrecision  int
	workbook        *xlsx.File
	sheet           *xlsx.Sheet
	output          *os.File
	closed          bool
	Info            map[string]Info
	elementOrder    []string
	formedAtomOrder []string
	TotalInfo       InfoWithCombinedAtoms
	ElapsedTime     float64
}

// Info - structure containing details about the simulation progress.
type Info struct {
	AtomsOnSurface int
	AdsorbedAtoms  int
	DesorbedAtoms  int
	Density        float64
	DensityF       float64
	DensityS       float64
	RecombEr       float64
	RecombLhF      float64
	RecombLhS      float64
}

type InfoWithCombinedAtoms struct {
	Info
	FormedAtoms map[string]int
}

// NewInfoCollector creates a new InfoCollector. It also generates an Excel file with a pre-filled header.
func NewInfoCollector(fileName string, floatPrecision int, elements []configs.Element, formedAtomNames []string) (*InfoCollector, error) {
	file := xlsx.NewFile()
	sh, err := file.AddSheet("Sheet1")
	if err != nil {
		return nil, err
	}

	headers := []string{
		"Simulation time",
	}

	// Add total info headers
	headers = append(headers,
		"Qty atoms on surface",
		"Qty adsorbed atoms",
		"Qty desorbed atoms",
		"Surface coverage",
		"Density F",
		"Density S",
		"Recomb Er",
		"Recomb Lh F",
		"Recomb Lh S",
	)

	for _, formedAtomName := range formedAtomNames {
		headers = append(headers, fmt.Sprintf("%s - Formed count", formedAtomName))
	}

	elementOrder := make([]string, 0, len(elements))
	if len(elements) > 1 {
		for _, element := range elements {
			headers = append(headers,
				fmt.Sprintf("%s - Qty atoms on surface", element.Name),
				fmt.Sprintf("%s - Qty adsorbed atoms", element.Name),
				fmt.Sprintf("%s - Qty desorbed atoms", element.Name),
				fmt.Sprintf("%s - Surface coverage", element.Name),
				fmt.Sprintf("%s - Density F", element.Name),
				fmt.Sprintf("%s - Density S", element.Name),
				fmt.Sprintf("%s - Recomb Er", element.Name),
				fmt.Sprintf("%s - Recomb Lh F", element.Name),
				fmt.Sprintf("%s - Recomb Lh S", element.Name),
			)
			elementOrder = append(elementOrder, element.Name)
		}
	}

	row := sh.AddRow()
	for _, header := range headers {
		row.AddCell().SetString(header)
	}
	row = sh.AddRow()
	for range len(headers) {
		row.AddCell().SetFloat(0)
	}

	output, err := os.Create(fileName)
	if err != nil {
		return nil, err
	}

	info := make(map[string]Info)
	for _, element := range elements {
		info[element.Name] = Info{}
	}

	collector := &InfoCollector{
		fileName:        fileName,
		floatPrecision:  floatPrecision,
		workbook:        file,
		sheet:           sh,
		output:          output,
		Info:            info,
		TotalInfo:       InfoWithCombinedAtoms{FormedAtoms: make(map[string]int)},
		elementOrder:    elementOrder,
		formedAtomOrder: formedAtomNames,
	}

	if err = collector.Flush(); err != nil {
		_ = output.Close()
		return nil, err
	}

	return collector, nil
}

// WriteInfo collects information about the simulation progress.
func (i *InfoCollector) WriteInfo() error {
	row := i.sheet.AddRow()

	// Write common info (step and time)
	row.AddCell().SetFloat(roundToDecimals(i.ElapsedTime, i.floatPrecision))

	// Write total info
	row.AddCell().SetInt(i.TotalInfo.AtomsOnSurface)
	row.AddCell().SetInt(i.TotalInfo.AdsorbedAtoms)
	row.AddCell().SetInt(i.TotalInfo.DesorbedAtoms)
	row.AddCell().SetFloat(roundToDecimals(i.TotalInfo.Density, i.floatPrecision))
	row.AddCell().SetFloat(roundToDecimals(i.TotalInfo.DensityF, i.floatPrecision))
	row.AddCell().SetFloat(roundToDecimals(i.TotalInfo.DensityS, i.floatPrecision))
	row.AddCell().SetFloat(roundToDecimals(i.TotalInfo.RecombEr, i.floatPrecision))
	row.AddCell().SetFloat(roundToDecimals(i.TotalInfo.RecombLhF, i.floatPrecision))
	row.AddCell().SetFloat(roundToDecimals(i.TotalInfo.RecombLhS, i.floatPrecision))
	for _, formedAtomName := range i.formedAtomOrder {
		row.AddCell().SetInt(i.TotalInfo.FormedAtoms[formedAtomName])
	}

	if len(i.elementOrder) < 2 {
		return i.Flush()
	}

	// Write element-specific info
	for _, element := range i.elementOrder {
		info := i.Info[element]
		row.AddCell().SetInt(info.AtomsOnSurface)
		row.AddCell().SetInt(info.AdsorbedAtoms)
		row.AddCell().SetInt(info.DesorbedAtoms)
		row.AddCell().SetFloat(roundToDecimals(info.Density, i.floatPrecision))
		row.AddCell().SetFloat(roundToDecimals(info.DensityF, i.floatPrecision))
		row.AddCell().SetFloat(roundToDecimals(info.DensityS, i.floatPrecision))
		row.AddCell().SetFloat(roundToDecimals(info.RecombEr, i.floatPrecision))
		row.AddCell().SetFloat(roundToDecimals(info.RecombLhF, i.floatPrecision))
		row.AddCell().SetFloat(roundToDecimals(info.RecombLhS, i.floatPrecision))
	}

	return i.Flush()
}

func (i *InfoCollector) Flush() error {
	if i.closed {
		return nil
	}

	if _, err := i.output.Seek(0, 0); err != nil {
		return err
	}
	if err := i.output.Truncate(0); err != nil {
		return err
	}
	if err := i.workbook.Write(i.output); err != nil {
		return err
	}

	return i.output.Sync()
}

func (i *InfoCollector) Close() error {
	if i.closed {
		return nil
	}

	err := i.Flush()
	closeErr := i.output.Close()
	i.closed = true
	if err != nil {
		return err
	}
	return closeErr
}

func roundToDecimals(value float64, precision int) float64 {
	factor := math.Pow(10, float64(precision))
	return math.Round(value*factor) / factor
}
