package simulator

import (
	"fmt"
	"log"
	"log/slog"
	"main/internal"
	"main/internal/config"
	"math/rand"
	"sort"
	"time"
)

type Simulator struct {
	cfg             config.Config
	matrix          *internal.Matrix
	atomsController *internal.SurfaceAtomsController
	infoCollector   *internal.InfoCollector
	temperature     int
	simulatingSteps int
	currentStep     int
	meta            SimulationMeta
}

func NewSimulator(cfg config.Config, temperature, simulatingSteps int) *Simulator {
	matrix := internal.NewMatrix(cfg.Constants)
	matrix.Init(cfg.Simulating.MatrixLenX, cfg.Simulating.MatrixLenY)

	atomsController := internal.NewSurfaceAtomsController(cfg.Simulating.MatrixLenX, cfg.Simulating.MatrixLenY, matrix)

	meta := Fill(cfg.Constants, float64(temperature))

	startTime := time.Now().Format("2006-01-02 15_04_05")
	infoCollector, err := internal.NewInfoCollector(fmt.Sprintf("result %s T%dK.xlsx", startTime, temperature))
	if err != nil {
		log.Fatal(err)
	}

	return &Simulator{
		cfg:             cfg,
		matrix:          matrix,
		atomsController: atomsController,
		temperature:     temperature,
		simulatingSteps: simulatingSteps,
		infoCollector:   infoCollector,
		meta:            meta,
	}
}

// Simulate - функция, которая моделирует процесс адсорбции, диффузии, рекомбинации и десорбции атомов на поверхности.
// Она использует алгоритм Монте-Карло, чтобы выбрать процесс, который будет происходить в следующем шаге.
// Затем она выбирает случайный атом, который будет участвовать в этом процессе.
// Если процесс - это адсорбция, то атом будет установлен на ячейку, если она свободна.
// Если процесс - это десорбция, то атом будет удален из ячейки, если он там был.
// Если процесс - это диффузия, то атом будет перемещен в случайную ячейку, если она свободна.
// Каждые 10% симуляции будет выводиться информация о ходе симуляции.
// Кроме того, каждые 10% симуляции будет записываться информация в excel файл.
func (s *Simulator) Simulate() {
	startTime := time.Now()
	progressModer := float64(s.simulatingSteps) * 0.1
	excelWriteModer := float64(s.simulatingSteps) * s.cfg.Simulating.LogPercent / 100

	for step := range s.simulatingSteps + 1 {
		s.currentStep = step
		if step%int(progressModer) == 0 || (step/int(progressModer)) == 0 && step%int(progressModer*0.1) == 0 {
			slog.Info(fmt.Sprintf("Simulated %d%%", step/int(progressModer*0.1)), "time", time.Since(startTime))
		}

		process, spendTime, randomNumber := s.getProcess()
		s.infoCollector.Info.ElapsedTime += spendTime

		switch process {
		case adsorptionSProcess:
			s.adsorbAtom('S')
		case adsorptionFProcess:
			s.adsorbAtom('F')
		case recombErProcess:
			s.desorbAtom('S')
			s.infoCollector.Info.DesorbedAtoms += 2
			s.infoCollector.Info.RecombEr += 1
		case desorptionFProcess:
			s.desorbAtom('F')
			s.infoCollector.Info.DesorbedAtoms += 1
		case diffusionProcess:
			desorbedAtoms := s.moveRandomAtom(randomNumber)
			s.infoCollector.Info.DesorbedAtoms += desorbedAtoms
		}

		if step%int(excelWriteModer) == 0 {
			s.infoCollector.Info.Step = step
			s.infoCollector.Info.AtomsOnSurface = len(s.atomsController.AtomsOnSurface)
			s.infoCollector.Info.Density = float64(len(s.atomsController.AtomsOnSurface)) / (float64(s.atomsController.MatrixLimitX) * float64(s.atomsController.MatrixLimitY))
			s.infoCollector.Info.DensityF = float64(s.atomsController.AtomsOnFCenters.Len()) / (float64(s.matrix.NumOfFSites))
			s.infoCollector.Info.DensityS = float64(s.atomsController.AtomsOnSCenters.Len()) / (float64(s.matrix.NumOfSSites))
			s.infoCollector.WriteInfo()
		}
	}
}

const (
	adsorptionFProcess = "adsorptionF"
	adsorptionSProcess = "adsorptionS"
	recombErProcess    = "recombEr"
	desorptionFProcess = "desorptionF"
	diffusionProcess   = "diffusion"
)

func (s *Simulator) getProcess() (process string, processTime float64, randomizedNumber float64) {
	lambdaAdsorptionF := s.calcLambdaAdsorptionF()
	lambdaAdsorptionS := s.calcLambdaAdsorptionS()
	lambdaRecombEr := s.calcLambdaRecombEr()
	lambdaDesorptionF := s.calcLambdaDesorptionF()
	lambdaDiffusions := s.calcLambdaDiffusion()

	lambda := lambdaAdsorptionF +
		lambdaAdsorptionS +
		lambdaRecombEr +
		lambdaDesorptionF +
		lambdaDiffusions

	action := map[float64]string{
		lambdaAdsorptionF / lambda: adsorptionFProcess,
		lambdaAdsorptionS / lambda: adsorptionSProcess,
		lambdaRecombEr / lambda:    recombErProcess,
		lambdaDesorptionF / lambda: desorptionFProcess,
		lambdaDiffusions / lambda:  diffusionProcess,
	}

	probabilityList := []float64{
		lambdaAdsorptionF / lambda,
		lambdaAdsorptionS / lambda,
		lambdaRecombEr / lambda,
		lambdaDesorptionF / lambda,
		lambdaDiffusions / lambda,
	}
	sort.Float64s(probabilityList)

	randomNumber := rand.Float64()

	spentTime := CalcTime(lambda, randomNumber)

	cumulativeProbability := 0.0
	for _, probability := range probabilityList {
		cumulativeProbability += probability
		if randomNumber <= cumulativeProbability {
			return action[probability], spentTime, randomNumber
		}
	}

	return "nothing", 0, randomNumber
}

func (s *Simulator) adsorbAtom(center rune) {
	var freeCells *internal.RandMap[int, internal.CellData]

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

	atom := internal.Atom{
		X:              cellData.X,
		Y:              cellData.Y,
		OccupiedCentre: cellData.Center,
	}
	s.atomsController.AddAtomOnSurface(atom)

	s.infoCollector.Info.AdsorbedAtoms++
}

func (s *Simulator) desorbAtom(center rune) {
	var atoms *internal.RandMap[int, internal.Atom]

	switch center {
	case 'S':
		atoms = s.atomsController.AtomsOnSCenters
	case 'F':
		atoms = s.atomsController.AtomsOnFCenters
	}

	cellId, atom, exist := atoms.Random()
	if !exist {
		slog.Error("no occupied cells", "cell_id", cellId)
	}

	s.atomsController.RemoveAtomFromSurface(atom.Id)
}

func (s *Simulator) moveRandomAtom(randomNumber float64) (desorbedAtoms int) {
	_, atom, exist := s.atomsController.AtomsOnFCenters.Random()
	if !exist {
		slog.Info("no atoms to move",
			"atoms_on_f_centers", s.atomsController.AtomsOnFCenters.Len(),
			"atoms_on_s_centers", s.atomsController.AtomsOnSCenters.Len(),
			"surface_atoms", len(s.atomsController.AtomsOnSurface))
		return
	}

	nextX, nextY := s.atomsController.GetNextAtomCoordinates(atom.Id)

	nextCellInfo := s.matrix.GetCellInfo(nextX, nextY)

	switch {
	case nextCellInfo.IsFree:
		s.atomsController.MoveAtom(atom, nextCellInfo)
	case nextCellInfo.Center == 'S' && s.meta.recombinationProbabilityOnSSite >= randomNumber:
		s.atomsController.RemoveAtomFromSurface(atom.Id)
		s.atomsController.RemoveAtomFromSurface(nextCellInfo.AtomId)

		desorbedAtoms = 2
		s.infoCollector.Info.RecombLhS += 1
	case nextCellInfo.Center == 'F' && s.meta.recombinationProbabilityOnFSite >= randomNumber:
		s.atomsController.RemoveAtomFromSurface(atom.Id)
		s.atomsController.RemoveAtomFromSurface(nextCellInfo.AtomId)

		desorbedAtoms = 2
		s.infoCollector.Info.RecombLhF += 1
	default:
		s.atomsController.RemoveAtomFromSurface(atom.Id)

		desorbedAtoms = 1
	}
	return
}
