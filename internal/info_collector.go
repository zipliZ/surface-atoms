package internal

import (
	"log"

	"github.com/tealeg/xlsx"
)

type InfoCollector struct {
	fileName string
	Info     Info
}

func NewInfoCollector(fileName string) (*InfoCollector, error) {
	file := xlsx.NewFile()
	sh, err := file.AddSheet("Лист1")
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
	}

	row := sh.AddRow()
	for _, header := range headers {
		row.AddCell().SetString(header)
	}

	// сохраняем файл в filePath
	if err = file.Save(fileName); err != nil {
		return nil, err
	}

	return &InfoCollector{
		fileName: fileName,
		Info:     Info{},
	}, nil
}

type Info struct {
	Step           int
	ElapsedTime    float64
	AtomsOnSurface int
	AdsorbedAtoms  int
	DesorbedAtoms  int
	Density        float64
	DensityF       float64
	DensityS       float64
}

func (i *InfoCollector) CollectInfo() {
	file, err := xlsx.OpenFile(i.fileName)
	if err != nil {
		log.Fatal(err)
	}

	row := file.Sheet["Лист1"].AddRow()
	row.AddCell().SetInt(i.Info.Step)
	row.AddCell().SetFloat(i.Info.ElapsedTime)
	row.AddCell().SetInt(i.Info.AtomsOnSurface)
	row.AddCell().SetInt(i.Info.AdsorbedAtoms)
	row.AddCell().SetInt(i.Info.DesorbedAtoms)
	row.AddCell().SetFloat(i.Info.Density)
	row.AddCell().SetFloat(i.Info.DensityF)
	row.AddCell().SetFloat(i.Info.DensityS)

	if err = file.Save(i.fileName); err != nil {
		log.Fatal(err)
	}
}
