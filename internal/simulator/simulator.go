package simulator

import (
	"fmt"
	"log"
	"log/slog"
	"main/configs"
	"main/internal/graphic_plotter"
	"main/internal/random"
	"os"
	"sort"
	"time"
)

type Simulator struct {
	cfg             configs.Config
	matrix          *Matrix
	atomsController *SurfaceAtomsController
	infoCollector   *InfoCollector
	temperature     int
	simulatingSteps int
	currentStep     int
	meta            map[string]SimulationMeta
	graphicPlotter  *graphic_plotter.GraphicPlotter
}

func NewSimulator(cfg configs.Config, temperature, simulatingSteps int) *Simulator {
	matrix := NewMatrix(cfg.Constants)
	matrix.Init(cfg.Simulating.MatrixLenX, cfg.Simulating.MatrixLenY)

	atomsController := NewSurfaceAtomsController(cfg.Simulating.MatrixLenX, cfg.Simulating.MatrixLenY, matrix, cfg.Elements)

	meta := make(map[string]SimulationMeta)
	for _, element := range cfg.Elements {
		meta[element.Name] = Fill(element, cfg.Constants, float64(temperature))
	}

	startTime := time.Now().Format("2006-01-02 15_04_05")
	dirName := fmt.Sprintf("result %s T%dK", startTime, temperature)
	err := os.Mkdir(dirName, 0755)
	if err != nil {
		log.Fatal(err)
	}

	excelFileName := fmt.Sprintf("result_%s_T%dK.xlsx", startTime, temperature)
	infoCollector, err := NewInfoCollector(
		dirName+string(os.PathSeparator)+excelFileName,
		cfg.Simulating.FloatPrecision,
		cfg.Elements)
	if err != nil {
		log.Fatal(err)
	}

	graphicsFileName := fmt.Sprintf("graphics_T%dK.html", temperature)
	graphicPlotter := graphic_plotter.New(
		dirName+string(os.PathSeparator)+excelFileName,
		dirName+string(os.PathSeparator)+graphicsFileName,
		fmt.Sprintf("T%dK", temperature),
		cfg.Simulating.GraphicsToPlot)

	return &Simulator{
		cfg:             cfg,
		matrix:          matrix,
		atomsController: atomsController,
		temperature:     temperature,
		simulatingSteps: simulatingSteps,
		infoCollector:   infoCollector,
		graphicPlotter:  graphicPlotter,
		meta:            meta,
	}
}

// Simulate - function that simulates the processes of adsorption, diffusion, recombination, and desorption of atoms on a surface.
// It uses the Monte Carlo algorithm to determine which process will occur in the next step.
// Then, it selects a random atom to participate in this process.
// If the process is adsorption, the atom will be placed in a cell if it is free.
// If the process is desorption, the atom will be removed from the cell if it was present.
// If the process is diffusion, the atom will move to a random cell if it is free.
// Every 10% of the simulation, progress information will be displayed.
// Additionally, every 10% of the simulation, data will be recorded in an Excel file.
func (s *Simulator) Simulate() {
	startTime := time.Now()
	progressModer := float64(s.simulatingSteps) * 0.1
	excelWriteModer := float64(s.simulatingSteps) * s.cfg.Simulating.LogPercent / 100

	for step := 1; step <= s.simulatingSteps; step++ {
		s.currentStep = step
		if step%int(progressModer) == 0 || (step/int(progressModer)) == 0 && step%int(progressModer*0.1) == 0 {
			slog.Info(fmt.Sprintf("Simulated %d%%", step/int(progressModer*0.1)), "time", time.Since(startTime))
		}

		process, elementName, spendTime := s.getProcess()
		s.infoCollector.ElapsedTime += spendTime

		switch process {
		case adsorptionSProcess:
			s.adsorbAtom('S', elementName)
		case adsorptionFProcess:
			s.adsorbAtom('F', elementName)
		case recombErProcess:
			s.RecombEr(elementName)
		case desorptionFProcess:
			s.desorbAtom('F', elementName)
		case diffusionProcess:
			s.moveRandomAtom(elementName, s.meta[elementName])
		}

		if step%int(excelWriteModer) == 0 {
			s.infoCollector.Step = step
			total := Info{}
			for elementName = range s.meta {
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

			s.infoCollector.WriteInfo()
		}
	}

	err := s.graphicPlotter.Plot()
	if err != nil {
		slog.Error("plot error", "err", err)
	}
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

	randomNumber := random.Float64()
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
	var freeCells *random.Map[int, CellData]

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
	var atoms *random.Map[int, Atom]

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

func (s *Simulator) RecombEr(elementName string) {
	s.desorbAtom('S', elementName)

	info := s.infoCollector.Info[elementName]
	info.RecombEr += 1
	info.DesorbedAtoms += 1
	s.infoCollector.Info[elementName] = info
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
	case nextCellInfo.Center == 'S' && meta.recombinationProbabilityOnSSite >= random.Float64():
		info.DesorbedAtoms += 1
		info.RecombLhS += 1
		s.infoCollector.Info[elementName] = info

		nextAtom := s.atomsController.AtomsOnSurface[nextCellInfo.AtomId]
		nextElementInfo := s.infoCollector.Info[nextAtom.ElementName]
		nextElementInfo.DesorbedAtoms += 1
		s.infoCollector.Info[nextAtom.ElementName] = nextElementInfo

		s.atomsController.RemoveAtomFromSurface(atom.Id)
		s.atomsController.RemoveAtomFromSurface(nextCellInfo.AtomId)
	case nextCellInfo.Center == 'F' && meta.recombinationProbabilityOnFSite >= random.Float64():
		info.DesorbedAtoms += 1
		info.RecombLhF += 1
		s.infoCollector.Info[elementName] = info

		nextAtom := s.atomsController.AtomsOnSurface[nextCellInfo.AtomId]
		nextElementInfo := s.infoCollector.Info[nextAtom.ElementName]
		nextElementInfo.DesorbedAtoms += 1
		s.infoCollector.Info[nextAtom.ElementName] = nextElementInfo

		s.atomsController.RemoveAtomFromSurface(atom.Id)
		s.atomsController.RemoveAtomFromSurface(nextCellInfo.AtomId)
	default:
		s.atomsController.RemoveAtomFromSurface(atom.Id)

		info.DesorbedAtoms += 1
		s.infoCollector.Info[elementName] = info
	}
	return
}
