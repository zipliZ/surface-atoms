package simulator

import (
	"math"
)

// calcProbabilityEr calculates the probability of recombination at the S-center.
func (s *Simulator) calcProbabilityEr() float64 {
	sigmaS := float64(s.atomsController.AtomsOnSCenters.Len()) / float64(s.matrix.NumOfSSites)
	probabilityEr := 2.0 * sigmaS * s.meta.r4 / s.meta.atomFlux * s.cfg.Constants.SDensity

	return probabilityEr
}

// calcProbabilitySLh calculates the probability of recombination at both the S-center and F-center.
func (s *Simulator) calcProbabilitySLh() float64 {
	sigmaS := float64(s.atomsController.AtomsOnSCenters.Len()) / float64(s.matrix.NumOfSSites)
	sigmaF := float64(s.atomsController.AtomsOnFCenters.Len()) / float64(s.matrix.NumOfFSites)

	probabilitySLh := 2.0 * sigmaS * sigmaF * s.meta.r6 * s.cfg.Constants.SDensity / s.meta.atomFlux

	return probabilitySLh
}

// calcProbabilityFLh calculates the probability of recombination at the F-center.
func (s *Simulator) calcProbabilityFLh() float64 {
	sigmaF := float64(s.atomsController.AtomsOnFCenters.Len()) / float64(s.matrix.NumOfFSites)

	probabilityFLh := 2.0 * sigmaF * s.meta.r7 * s.cfg.Constants.FDensity / s.meta.atomFlux

	return probabilityFLh
}

// calcLambdaAdsorptionF calculates the adsorption rate at the F-center.
func (s *Simulator) calcLambdaAdsorptionF() float64 {
	lambdaAdsorptions := float64(s.matrix.CountFreeCellsOfFCenters()) * s.meta.atomFlux / (s.cfg.Constants.FDensity + s.cfg.Constants.SDensity)

	return lambdaAdsorptions
}

// calcLambdaAdsorptionS calculates the adsorption rate at the S-center.
func (s *Simulator) calcLambdaAdsorptionS() float64 {
	lambdaAdsorptions := float64(s.matrix.CountFreeCellsOfSCenters()) * s.meta.atomFlux / (s.cfg.Constants.FDensity + s.cfg.Constants.SDensity)

	return lambdaAdsorptions
}

// calcLambdaDesorptionF calculates the desorption rate at the F-center.
func (s *Simulator) calcLambdaDesorptionF() float64 {
	lambdaDesorption := float64(s.atomsController.AtomsOnFCenters.Len()) * s.meta.r2

	return lambdaDesorption
}

// calcLambdaDiffusion calculates the diffusion rate at the F-center.
func (s *Simulator) calcLambdaDiffusion() float64 {
	lambdaDiffusion := float64(s.atomsController.AtomsOnFCenters.Len()) * s.meta.r5
	return lambdaDiffusion
}

// calcLambdaRecombEr calculates the recombination rate at the S-center.
func (s *Simulator) calcLambdaRecombEr() float64 {
	lambdaRecombEr := float64(s.atomsController.AtomsOnSCenters.Len()) * s.meta.r4

	return lambdaRecombEr
}

// CalcTime calculates the physical time.
func CalcTime(lambda float64, randomNumber float64) float64 {
	return 1.0 / lambda * math.Log(1.0/randomNumber)
}
