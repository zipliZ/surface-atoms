package simulator

import (
	"fmt"
	"log"
	"main/configs"
	"math"

	"github.com/tealeg/xlsx"
)

// InfoCollector - structure that collects information about the simulation progress.
type InfoCollector struct {
	fileName       string
	floatPrecision int
	Info           map[string]Info
	elementOrder   []string
	TotalInfo      Info
	ElapsedTime    float64
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

// NewInfoCollector creates a new InfoCollector. It also generates an Excel file with a pre-filled header.
func NewInfoCollector(fileName string, floatPrecision int, elements []configs.Element) (*InfoCollector, error) {
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

	// Save the file to filePath
	if err = file.Save(fileName); err != nil {
		return nil, err
	}

	info := make(map[string]Info)
	for _, element := range elements {
		info[element.Name] = Info{}
	}

	return &InfoCollector{
		fileName:       fileName,
		floatPrecision: floatPrecision,
		Info:           info,
		TotalInfo:      Info{},
		elementOrder:   elementOrder,
	}, nil
}

// WriteInfo collects information about the simulation progress.
func (i *InfoCollector) WriteInfo() {
	file, err := xlsx.OpenFile(i.fileName)
	if err != nil {
		log.Fatal(err)
	}

	row := file.Sheet["Sheet1"].AddRow()

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

	if len(i.elementOrder) < 2 {
		if err = file.Save(i.fileName); err != nil {
			log.Fatal(err)
		}
		return
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

	if err = file.Save(i.fileName); err != nil {
		log.Fatal(err)
	}
}

func roundToDecimals(value float64, precision int) float64 {
	factor := math.Pow(10, float64(precision))
	return math.Round(value*factor) / factor
}
