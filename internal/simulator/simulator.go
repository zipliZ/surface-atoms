package simulator

import (
	"fmt"
	"log"
	"log/slog"
	"main/internal"
	"math/rand"
	"sort"
	"time"
)

type Simulator struct {
	matrix          *internal.Matrix
	atomsController *internal.SurfaceAtomsController
	infoCollector   *internal.InfoCollector
	temperature     int
	simulatingSteps int
	currentStep     int
	meta            SimulationMeta
}

func NewSimulator(temperature, simulatingSteps, matrixSizeX, matrixSizeY int) *Simulator {
	matrix := internal.NewMatrix()
	matrix.Init(matrixSizeX, matrixSizeY)

	atomsController := internal.NewSurfaceAtomsController(matrixSizeX, matrixSizeY, matrix)

	meta := Fill(float64(temperature))

	startTime := time.Now().Format("2006-01-02 15_04_05")
	infoCollector, err := internal.NewInfoCollector(startTime + ".xlsx")
	if err != nil {
		log.Fatal(err)
	}

	return &Simulator{
		matrix:          matrix,
		atomsController: atomsController,
		temperature:     temperature,
		simulatingSteps: simulatingSteps,
		infoCollector:   infoCollector,
		meta:            meta,
	}
}

func (s *Simulator) Simulate() {
	startTime := time.Now()
	for step := range s.simulatingSteps + 1 {
		s.currentStep = step
		moder := float64(s.simulatingSteps) * 0.1
		if step%int(moder) == 0 || (step/int(moder)) == 0 && step%int(moder*0.1) == 0 {
			slog.Info(fmt.Sprintf("Simulated %d%%", step/int(moder*0.1)), "time", time.Since(startTime))

			s.infoCollector.Info.Step = step
			s.infoCollector.Info.AtomsOnSurface = len(s.atomsController.AtomsOnSurface)
			s.infoCollector.Info.Density = float64(len(s.atomsController.AtomsOnSurface)) / (float64(s.atomsController.MatrixLimitX) * float64(s.atomsController.MatrixLimitY))
			s.infoCollector.Info.DensityF = float64(s.atomsController.AtomsOnFCenters.Len()) / (float64(s.matrix.NumOfFSites))
			s.infoCollector.Info.DensityS = float64(s.atomsController.AtomsOnSCenters.Len()) / (float64(s.matrix.NumOfSSites))
			s.infoCollector.CollectInfo()
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
		case desorptionFProcess:
			s.desorbAtom('F')
		case diffusionProcess:
			s.moveRandomAtom(randomNumber)
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

	spentTime := CalcTime(lambda)

	for _, probability := range probabilityList {
		if randomNumber <= probability {
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

	s.infoCollector.Info.DesorbedAtoms++
}

func (s *Simulator) moveRandomAtom(randomNumber float64) {
	_, atom, exist := s.atomsController.AtomsOnFCenters.Random()
	if !exist {
		slog.Info("no atoms to move", "atoms_on_f_centers", s.atomsController.AtomsOnFCenters.Len())
		return
	}

	nextX, nextY := s.atomsController.GetNextAtomCoordinates(atom.Id)

	nextCellInfo := s.matrix.GetCellInfo(nextX, nextY)

	switch {
	case nextCellInfo.IsFree:
		s.atomsController.MoveAtom(atom, nextCellInfo)
	case nextCellInfo.Center == 'S' && s.meta.recombinationProbabilityOnFSite >= randomNumber:
		s.atomsController.RemoveAtomFromSurface(atom.Id)
		s.atomsController.RemoveAtomFromSurface(nextCellInfo.AtomId)

		s.infoCollector.Info.DesorbedAtoms += 2
	case nextCellInfo.Center == 'F' && s.meta.recombinationProbabilityOnFSite >= randomNumber:
		s.atomsController.RemoveAtomFromSurface(atom.Id)
		s.atomsController.RemoveAtomFromSurface(nextCellInfo.AtomId)

		s.infoCollector.Info.DesorbedAtoms += 2
	}
}
