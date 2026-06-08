package simulation

import (
	"cmp"
	"container/list"
	"fmt"
	"log/slog"
	"main/configs"
	"main/internal/graphic_plotter"
	randomx "main/internal/random"
	"math"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"
)

type Simulator struct {
	cfg                   configs.Config
	matrix                *Matrix
	atomsController       *SurfaceAtomsController
	infoCollector         *InfoCollector
	temperature           int
	simulationTime        float64
	currentSimulationTime float64
	meta                  map[string]SimulationMeta
	elems                 []string
	elementsByName        map[string]configs.Element
	graphicPlotter        *graphic_plotter.GraphicPlotter

	// [elementName][parameterName]Values
	elementValues         map[string]map[string]*Values
	stableIterationsCount int
}

type Values struct {
	Values *list.List
	Total  float64
}

func NewValues() *Values {
	return &Values{
		Values: list.New(),
		Total:  0,
	}
}

func NewSimulator(cfg configs.Config, temperature int, simulationTime float64) (*Simulator, error) {
	matrix := NewMatrix(cfg.Constants)
	matrix.Init(cfg.Simulating.MatrixLenX, cfg.Simulating.MatrixLenY)

	atomsController := NewSurfaceAtomsController(cfg.Simulating.MatrixLenX, cfg.Simulating.MatrixLenY, matrix, cfg.Elements)

	var (
		meta           = make(map[string]SimulationMeta)
		elems          = make([]string, 0, len(cfg.Elements))
		elementsByName = make(map[string]configs.Element, len(cfg.Elements))
	)

	for _, element := range cfg.Elements {
		meta[element.Name] = Fill(element, cfg.Constants, float64(temperature))
		elems = append(elems, element.Name)
		elementsByName[element.Name] = element
	}

	startTime := time.Now().Format("2006-01-02 15_04_05")
	dirName := fmt.Sprintf("result %s T%dK", startTime, temperature)
	err := os.Mkdir(dirName, 0755)
	if err != nil {
		return nil, err
	}

	excelFileName := fmt.Sprintf("result_%s_T%dK.xlsx", startTime, temperature)
	infoCollector, err := NewInfoCollector(
		filepath.Join(dirName, excelFileName),
		cfg.Simulating.FloatPrecision,
		cfg.Elements,
		GetFormedAtomNames(cfg.Elements),
	)
	if err != nil {
		return nil, err
	}

	graphicsFileName := fmt.Sprintf("result_%s_T%dK.html", startTime, temperature)
	graphicPlotter := graphic_plotter.New(
		filepath.Join(dirName, excelFileName),
		filepath.Join(dirName, graphicsFileName),
		fmt.Sprintf("T%dK", temperature),
		cfg.Simulating.GraphicsToPlot)

	return &Simulator{
		cfg:                   cfg,
		matrix:                matrix,
		atomsController:       atomsController,
		temperature:           temperature,
		simulationTime:        simulationTime,
		infoCollector:         infoCollector,
		graphicPlotter:        graphicPlotter,
		meta:                  meta,
		elems:                 elems,
		elementsByName:        elementsByName,
		elementValues:         make(map[string]map[string]*Values),
		stableIterationsCount: 0,
	}, nil
}

func GetCombinedAtomName(elements []configs.Element) string {
	sortedElements := slices.Clone(elements)
	slices.SortFunc(sortedElements, func(a, b configs.Element) int {
		if result := cmp.Compare(a.Electronegativity, b.Electronegativity); result != 0 {
			return result
		}
		return cmp.Compare(a.Name, b.Name)
	})

	names := make([]string, len(sortedElements))
	for i, element := range sortedElements {
		names[i] = element.Name
	}

	return strings.Join(names, "")
}

func GetFormedAtomName(first, second configs.Element) string {
	if first.Name == second.Name {
		return fmt.Sprintf("%s2", first.Name)
	}
	return GetCombinedAtomName([]configs.Element{first, second})
}

func GetFormedAtomNames(elements []configs.Element) []string {
	if len(elements) == 0 {
		return nil
	}

	names := make([]string, 0, len(elements)*len(elements))
	seen := make(map[string]struct{})

	if len(elements) > 1 {
		for i := 0; i < len(elements); i++ {
			for j := i + 1; j < len(elements); j++ {
				name := GetFormedAtomName(elements[i], elements[j])
				if _, exists := seen[name]; !exists {
					names = append(names, name)
					seen[name] = struct{}{}
				}
			}
		}
	}

	for _, element := range elements {
		name := GetFormedAtomName(element, element)
		if _, exists := seen[name]; !exists {
			names = append(names, name)
			seen[name] = struct{}{}
		}
	}

	return names
}

// Simulate - function that simulates the processes of adsorption, diffusion, recombination, and desorption of atoms on a surface.
// It uses the Monte Carlo algorithm to determine which process will occur in the next step.
// Then, it selects a randomx atom to participate in this process.
// If the process is adsorption, the atom will be placed in a cell if it is free.
// If the process is desorption, the atom will be removed from the cell if it was present.
// If the process is diffusion, the atom will move to a randomx cell if it is free.
// Every 10% of the simulation, progress information will be displayed.
// Additionally, every 10% of the simulation, data will be recorded in an Excel file.
func (s *Simulator) Simulate() (err error) {
	defer func() {
		if closeErr := s.infoCollector.Close(); err == nil && closeErr != nil {
			err = closeErr
		}
	}()

	startTime := time.Now()

	progressInterval := s.simulationTime * 0.1
	excelWriteInterval := s.simulationTime * s.cfg.Simulating.LogPercent / 100

	nextProgressTime := progressInterval
	nextExcelWriteTime := excelWriteInterval

	progressCount := 1
	for s.currentSimulationTime <= s.simulationTime {
		if s.currentSimulationTime >= nextProgressTime && progressCount <= 10 {
			currentPercent := progressCount * 10
			slog.Info(fmt.Sprintf("Simulated %d%%", currentPercent), "physical time", s.currentSimulationTime, "time", time.Since(startTime))
			nextProgressTime += progressInterval
			progressCount++
		}

		process, elementName, spendTime := s.getProcess()
		s.currentSimulationTime += spendTime
		s.infoCollector.ElapsedTime += spendTime

		switch process {
		case adsorptionSProcess:
			s.adsorbAtom('S', elementName)
		case adsorptionFProcess:
			s.adsorbAtom('F', elementName)
		case recombErProcess:
			s.recombEr(elementName)
		case desorptionFProcess:
			s.desorbAtom('F', elementName)
		case diffusionProcess:
			s.moveRandomAtom(elementName, s.meta[elementName])
		}

		if s.currentSimulationTime >= nextExcelWriteTime {
			if err = s.writeInfoSnapshot(); err != nil {
				return err
			}

			if s.checkQuasiSteadyState() {
				slog.Info("Quasi-steady state reached",
					"physical_time", s.currentSimulationTime,
					"elapsed_time", time.Since(startTime),
					"stable_iterations", s.stableIterationsCount,
					"checked_parameters", s.cfg.Simulating.CheckParameters)
				break
			}

			nextExcelWriteTime += excelWriteInterval
		}
	}

	if err = s.infoCollector.Close(); err != nil {
		return err
	}

	if err = s.graphicPlotter.Plot(); err != nil {
		slog.Error("plot error", "err", err)
		return err
	}

	return nil
}

func (s *Simulator) writeInfoSnapshot() error {
	s.infoCollector.ElapsedTime = s.currentSimulationTime
	total := InfoWithCombinedAtoms{
		FormedAtoms: make(map[string]int, len(s.infoCollector.TotalInfo.FormedAtoms)),
	}
	for formedAtomName, count := range s.infoCollector.TotalInfo.FormedAtoms {
		total.FormedAtoms[formedAtomName] = count
	}

	for elementName := range s.meta {
		info := s.infoCollector.Info[elementName]
		info.AtomsOnSurface = s.atomsController.AtomsOnFCenters[elementName].Len() + s.atomsController.AtomsOnSCenters[elementName].Len()
		info.Density = float64(info.AtomsOnSurface) / (float64(s.atomsController.MatrixLimitX) * float64(s.atomsController.MatrixLimitY))
		info.DensityF = float64(s.atomsController.AtomsOnFCenters[elementName].Len()) / (float64(s.matrix.NumOfFSites))
		info.DensityS = float64(s.atomsController.AtomsOnSCenters[elementName].Len()) / (float64(s.matrix.NumOfSSites))

		s.infoCollector.Info[elementName] = info

		total.AtomsOnSurface += info.AtomsOnSurface
		total.AdsorbedAtoms += info.AdsorbedAtoms
		total.DesorbedAtoms += info.DesorbedAtoms
		total.RecombEr += info.RecombEr
		total.RecombLhF += info.RecombLhF
		total.RecombLhS += info.RecombLhS
	}

	total.Density = float64(len(s.atomsController.AtomsOnSurface)) / (float64(s.atomsController.MatrixLimitX) * float64(s.atomsController.MatrixLimitY))
	total.DensityF = float64(s.atomsController.AtomsOnFCenters.Len()) / (float64(s.matrix.NumOfFSites))
	total.DensityS = float64(s.atomsController.AtomsOnSCenters.Len()) / (float64(s.matrix.NumOfSSites))
	s.infoCollector.TotalInfo = total

	return s.infoCollector.WriteInfo()
}

const (
	adsorptionFProcess = "adsorptionF"
	adsorptionSProcess = "adsorptionS"
	recombErProcess    = "recombEr"
	desorptionFProcess = "desorptionF"
	diffusionProcess   = "diffusion"
)

func (s *Simulator) getProcess() (process string, elementName string, processTime float64) {
	// Calculate total lambda for all elements and processes
	totalLambda := 0.0
	type processInfo struct {
		probability float64
		elementName string
		process     string
	}
	processes := make([]processInfo, 0)

	for name, meta := range s.meta {
		lambdaAdsorptionF := s.calcLambdaAdsorptionF(meta)
		lambdaAdsorptionS := s.calcLambdaAdsorptionS(meta)
		lambdaRecombEr := s.calcLambdaRecombEr(name, meta)
		lambdaDesorptionF := s.calcLambdaDesorptionF(name, meta)
		lambdaDiffusions := s.calcLambdaDiffusion(name, meta)

		// Add all lambdas to total
		totalLambda += lambdaAdsorptionF +
			lambdaAdsorptionS +
			lambdaRecombEr +
			lambdaDesorptionF +
			lambdaDiffusions

		// Store all processes with their lambdas
		processes = append(processes,
			processInfo{lambdaAdsorptionF, name, adsorptionFProcess},
			processInfo{lambdaAdsorptionS, name, adsorptionSProcess},
			processInfo{lambdaRecombEr, name, recombErProcess},
			processInfo{lambdaDesorptionF, name, desorptionFProcess},
			processInfo{lambdaDiffusions, name, diffusionProcess},
		)
	}

	// Calculate probabilities relative to total lambda
	for i := range processes {
		processes[i].probability = processes[i].probability / totalLambda
	}

	// Sort processes by probability
	sort.Slice(processes, func(i, j int) bool {
		return processes[i].probability < processes[j].probability
	})

	randomNumber := randomx.Float64()
	spentTime := CalcTime(totalLambda)

	cumulativeProbability := 0.0
	for _, proc := range processes {
		cumulativeProbability += proc.probability
		if randomNumber <= cumulativeProbability {
			return proc.process, proc.elementName, spentTime
		}
	}

	return "nothing", "", 0
}

func (s *Simulator) adsorbAtom(center rune, elementName string) {
	var freeCells *randomx.Map[uint32, CellData]

	switch center {
	case 'S':
		freeCells = s.matrix.FreeCellsOfSCenters
	case 'F':
		freeCells = s.matrix.FreeCellsOfFCenters
	}

	cellId, cellData, exist := freeCells.Random()
	if !exist {
		slog.Error("no free cells", "cell_id", cellId)
	}

	atom := Atom{
		X:              cellData.X,
		Y:              cellData.Y,
		OccupiedCentre: cellData.Center,
		ElementName:    elementName,
	}

	info := s.infoCollector.Info[elementName]
	info.AdsorbedAtoms += 1
	s.infoCollector.Info[elementName] = info

	s.atomsController.AddAtomOnSurface(atom)
}

func (s *Simulator) desorbAtom(center rune, elementName string) {
	var atoms *randomx.Map[int, Atom]

	switch center {
	case 'S':
		atoms = s.atomsController.AtomsOnSCenters[elementName]
	case 'F':
		atoms = s.atomsController.AtomsOnFCenters[elementName]
	}

	cellId, atom, exist := atoms.Random()
	if !exist {
		slog.Error("no occupied cells", "cell_id", cellId)
	}

	info := s.infoCollector.Info[elementName]
	info.DesorbedAtoms += 1
	s.infoCollector.Info[elementName] = info

	s.atomsController.RemoveAtomFromSurface(atom.Id)
}

func (s *Simulator) recombEr(elementName string) {
	info := s.infoCollector.Info[elementName]
	info.RecombEr += 1
	info.DesorbedAtoms += 1
	s.infoCollector.Info[elementName] = info

	randomElement := s.elems[randomx.Int(len(s.elems))]
	randomElementInfo := s.infoCollector.Info[randomElement]
	randomElementInfo.RecombEr += 1
	s.infoCollector.Info[randomElement] = randomElementInfo

	s.recordFormedAtom(elementName, randomElement)
	s.desorbAtom('S', randomElement)
}

func (s *Simulator) moveRandomAtom(elementName string, meta SimulationMeta) {
	_, atom, exist := s.atomsController.AtomsOnFCenters[elementName].Random()
	if !exist {
		slog.Info("no atoms to move",
			"element_name", elementName,
			"atoms_on_f_centers", s.atomsController.AtomsOnFCenters[elementName].Len(),
			"atoms_on_s_centers", s.atomsController.AtomsOnSCenters[elementName].Len(),
			"surface_atoms", len(s.atomsController.AtomsOnSurface))
		return
	}

	nextX, nextY := s.atomsController.GetNextAtomCoordinates(atom.Id)

	nextCellInfo := s.matrix.GetCellInfo(nextX, nextY)

	info := s.infoCollector.Info[elementName]
	switch {
	case nextCellInfo.IsFree:
		s.atomsController.MoveAtom(atom, nextCellInfo)
	case nextCellInfo.Center == 'S' && recombProbOnS(elementName, s.atomsController.AtomsOnSurface[nextCellInfo.AtomId].ElementName, meta) >= randomx.Float64():
		info.DesorbedAtoms += 1
		info.RecombLhS += 1
		s.infoCollector.Info[elementName] = info

		nextAtom := s.atomsController.AtomsOnSurface[nextCellInfo.AtomId]
		nextElementInfo := s.infoCollector.Info[nextAtom.ElementName]
		nextElementInfo.DesorbedAtoms += 1
		nextElementInfo.RecombLhS += 1
		s.infoCollector.Info[nextAtom.ElementName] = nextElementInfo

		s.recordFormedAtom(elementName, nextAtom.ElementName)

		s.atomsController.RemoveAtomFromSurface(atom.Id)
		s.atomsController.RemoveAtomFromSurface(nextCellInfo.AtomId)
	case nextCellInfo.Center == 'F' && recombProbOnF(elementName, s.atomsController.AtomsOnSurface[nextCellInfo.AtomId].ElementName, meta) >= randomx.Float64():
		info.DesorbedAtoms += 1
		info.RecombLhF += 1
		s.infoCollector.Info[elementName] = info

		nextAtom := s.atomsController.AtomsOnSurface[nextCellInfo.AtomId]
		nextElementInfo := s.infoCollector.Info[nextAtom.ElementName]
		nextElementInfo.DesorbedAtoms += 1
		nextElementInfo.RecombLhF += 1
		s.infoCollector.Info[nextAtom.ElementName] = nextElementInfo

		s.recordFormedAtom(elementName, nextAtom.ElementName)

		s.atomsController.RemoveAtomFromSurface(atom.Id)
		s.atomsController.RemoveAtomFromSurface(nextCellInfo.AtomId)
	default:
		s.atomsController.RemoveAtomFromSurface(atom.Id)

		info.DesorbedAtoms += 1
		s.infoCollector.Info[elementName] = info
	}
	return
}

func IsDifferentAtoms(a string, b string) bool {
	return a != b
}

func recombProbOnS(elementName, nextElementName string, meta SimulationMeta) float64 {
	if IsDifferentAtoms(elementName, nextElementName) {
		return meta.recombinationProbabilityOnSSiteHet
	}
	return meta.recombinationProbabilityOnSSite
}

func recombProbOnF(elementName, nextElementName string, meta SimulationMeta) float64 {
	if IsDifferentAtoms(elementName, nextElementName) {
		return meta.recombinationProbabilityOnFSiteHet
	}
	return meta.recombinationProbabilityOnFSite
}

func (s *Simulator) recordFormedAtom(firstElementName, secondElementName string) {
	firstElement, firstExists := s.elementsByName[firstElementName]
	secondElement, secondExists := s.elementsByName[secondElementName]
	if !firstExists || !secondExists {
		slog.Warn("formed atom contains unknown element",
			"first_element", firstElementName,
			"second_element", secondElementName)
		return
	}

	if s.infoCollector.TotalInfo.FormedAtoms == nil {
		s.infoCollector.TotalInfo.FormedAtoms = make(map[string]int)
	}

	formedAtomName := GetFormedAtomName(firstElement, secondElement)
	s.infoCollector.TotalInfo.FormedAtoms[formedAtomName]++
}

func (s *Simulator) checkQuasiSteadyState() bool {
	if !s.cfg.Simulating.StopOnQuasiSteady || len(s.cfg.Simulating.CheckParameters) == 0 {
		return false
	}

	isStable := true
	for elementName := range s.meta {
		info := s.infoCollector.Info[elementName]

		if s.elementValues[elementName] == nil {
			s.elementValues[elementName] = make(map[string]*Values)
		}

		for _, param := range s.cfg.Simulating.CheckParameters {
			var currentValue float64

			switch param.Name {
			case "density":
				currentValue = info.Density
			case "densityF":
				currentValue = info.DensityF
			case "densityS":
				currentValue = info.DensityS
			case "atomsOnSurface":
				currentValue = float64(info.AtomsOnSurface)
			default:
				slog.Warn("unknown parameter for quasi-steady check", "parameter", param)
				continue
			}

			values, exists := s.elementValues[elementName][param.Name]
			if !exists {
				s.elementValues[elementName][param.Name] = NewValues()
				isStable = false
				continue
			}

			var relativeChange float64
			if values.Values.Len() == param.ValuesWindowSize {
				meanValue := values.Total / float64(values.Values.Len())
				relativeChange = math.Abs((currentValue - meanValue) / meanValue)

				// Remove oldest value
				oldestElement := values.Values.Back()
				if oldestElement != nil {
					oldestValue := oldestElement.Value.(float64)
					values.Total -= oldestValue
					values.Values.Remove(oldestElement)
				}
				values.Values.PushFront(currentValue)
				values.Total += currentValue
				s.elementValues[elementName][param.Name] = values

			} else {
				relativeChange = 1.0
				values.Values.PushFront(currentValue)
				values.Total += currentValue
				s.elementValues[elementName][param.Name] = values
			}

			if relativeChange > param.Tolerance/100.0 {
				isStable = false
			}

		}
	}

	if isStable {
		s.stableIterationsCount++
	} else {
		s.stableIterationsCount = 0
	}

	return s.stableIterationsCount >= s.cfg.Simulating.RequiredStableChecks
}
