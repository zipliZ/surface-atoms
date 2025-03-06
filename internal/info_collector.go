package internal

import (
	"log"
	"math"

	"github.com/tealeg/xlsx"
)

// InfoCollector - structure that collects information about the simulation progress.
type InfoCollector struct {
	fileName       string
	floatPrecision int
	Info           Info
}

// Info - structure containing details about the simulation progress.
type Info struct {
	Step           int
	ElapsedTime    float64
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

// NewInfoCollector creates a new InfoCollector. It also generates an Excel file with a pre-filled header.
func NewInfoCollector(fileName string, floatPrecision int) (*InfoCollector, error) {
	file := xlsx.NewFile()
	sh, err := file.AddSheet("Sheet1")
	if err != nil {
		return nil, err
	}

	headers := []string{
		"Step N",
		"Simulation time",
		"Qty atoms on surface",
		"Qty adsorbed atoms",
		"Qty desorbed atoms",
		"Surface coverage",
		"Density F",
		"Density S",
		"Recomb Er",
		"Recomb Lh F",
		"Recomb Lh S",
	}

	row := sh.AddRow()
	for _, header := range headers {
		row.AddCell().SetString(header)
	}
	row = sh.AddRow()
	for range len(headers) {
		row.AddCell().SetFloat(0)
	}

	// Save the file to filePath
	if err = file.Save(fileName); err != nil {
		return nil, err
	}

	return &InfoCollector{
		fileName:       fileName,
		floatPrecision: floatPrecision,
		Info:           Info{},
	}, nil
}

// WriteInfo collects information about the simulation progress.
func (i *InfoCollector) WriteInfo() {
	file, err := xlsx.OpenFile(i.fileName)
	if err != nil {
		log.Fatal(err)
	}

	row := file.Sheet["Sheet1"].AddRow()
	row.AddCell().SetInt(i.Info.Step)
	row.AddCell().SetFloat(roundToDecimals(i.Info.ElapsedTime, i.floatPrecision))
	row.AddCell().SetInt(i.Info.AtomsOnSurface)
	row.AddCell().SetInt(i.Info.AdsorbedAtoms)
	row.AddCell().SetInt(i.Info.DesorbedAtoms)
	row.AddCell().SetFloat(roundToDecimals(i.Info.Density, i.floatPrecision))
	row.AddCell().SetFloat(roundToDecimals(i.Info.DensityF, i.floatPrecision))
	row.AddCell().SetFloat(roundToDecimals(i.Info.DensityS, i.floatPrecision))
	row.AddCell().SetFloat(roundToDecimals(i.Info.RecombEr, i.floatPrecision))
	row.AddCell().SetFloat(roundToDecimals(i.Info.RecombLhF, i.floatPrecision))
	row.AddCell().SetFloat(roundToDecimals(i.Info.RecombLhS, i.floatPrecision))

	if err = file.Save(i.fileName); err != nil {
		log.Fatal(err)
	}
}

func roundToDecimals(value float64, precision int) float64 {
	factor := math.Pow(10, float64(precision))
	return math.Round(value*factor) / factor
}
