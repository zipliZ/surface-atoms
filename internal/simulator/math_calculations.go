package simulator

import (
	"main/internal"
	"math"
	"math/rand/v2"
)

func (s *Simulator) calcProbabilityEr() float64 {
	sigmaS := float64(s.atomsController.AtomsOnSCenters.Len()) / float64(s.matrix.NumOfSSites)
	probabilityEr := 2.0 * sigmaS * s.meta.r4 / s.meta.atomFlux * internal.SDensity

	return probabilityEr
}

func (s *Simulator) calcProbabilitySLh() float64 {
	sigmaS := float64(s.atomsController.AtomsOnSCenters.Len()) / float64(s.matrix.NumOfSSites)
	sigmaF := float64(s.atomsController.AtomsOnFCenters.Len()) / float64(s.matrix.NumOfFSites)

	probabilitySLh := 2.0 * sigmaS * sigmaF * s.meta.r6 * internal.SDensity / s.meta.atomFlux

	return probabilitySLh
}

func (s *Simulator) calcProbabilityFLh() float64 {
	sigmaF := float64(s.atomsController.AtomsOnFCenters.Len()) / float64(s.matrix.NumOfFSites)

	probabilityFLh := 2.0 * sigmaF * s.meta.r7 * internal.FDensity / s.meta.atomFlux

	return probabilityFLh
}

func (s *Simulator) calcLambdaAdsorptionF() float64 {
	lambdaAdsorptions := float64(s.matrix.CountFreeCellsOfFCenters()) * s.meta.atomFlux / (internal.FDensity + internal.SDensity)

	return lambdaAdsorptions
}

func (s *Simulator) calcLambdaAdsorptionS() float64 {
	lambdaAdsorptions := float64(s.matrix.CountFreeCellsOFSCenters()) * s.meta.atomFlux / (internal.FDensity + internal.SDensity)

	return lambdaAdsorptions
}

func (s *Simulator) calcLambdaDesorptionF() float64 {
	lambdaDesorption := float64(s.atomsController.AtomsOnFCenters.Len()) * s.meta.r2

	return lambdaDesorption
}

func (s *Simulator) calcLambdaDiffusion() float64 {
	lambdaDesorption := float64(s.atomsController.AtomsOnFCenters.Len()+s.atomsController.AtomsOnSCenters.Len()) * s.meta.r5
	return lambdaDesorption
}

func (s *Simulator) calcLambdaRecombEr() float64 {
	lambdaRememberEr := float64(s.atomsController.AtomsOnFCenters.Len()) * s.meta.r4

	return lambdaRememberEr
}

// TODO Узнать используем 1 рандомное число или нет
func CalcTime(lambda float64) float64 {
	randNumber := rand.Float64()
	if randNumber == 0 {
		randNumber = 0.0000000000000000001
	}
	return 1.0 / lambda * math.Log(1.0/randNumber)
}
