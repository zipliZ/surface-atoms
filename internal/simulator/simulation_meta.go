package simulator

import (
	"main/internal/config"
	"math"
)

type SimulationMeta struct {
	atomFlux                        float64
	r1                              float64
	r2                              float64
	r3                              float64
	r4                              float64
	r5                              float64
	r6                              float64
	r7                              float64
	recombinationProbabilityOnSSite float64
	recombinationProbabilityOnFSite float64
}

func Fill(constants config.Constants, temperature float64) SimulationMeta {
	atomFlux := calculateAtomFlux(constants, temperature)
	r1 := calcR1(constants, atomFlux)
	r2 := calcR2(constants, temperature)
	r3 := calcR3(constants, atomFlux)
	r4 := calcR4(constants, temperature, r3)
	r5 := calcR5(constants, temperature)
	r6 := calcR6(constants, temperature, r5)
	r7 := calcR7(constants, temperature, r5)

	recombinationProbabilityOnSSite := calcRecombinationProbabilityOnSSite(constants, temperature)
	recombinationProbabilityOnFSite := calcRecombinationProbabilityOnFSite(constants, temperature)

	return SimulationMeta{
		atomFlux: atomFlux,
		r1:       r1,
		r2:       r2,
		r3:       r3,
		r4:       r4,
		r5:       r5,
		r6:       r6,
		r7:       r7,

		recombinationProbabilityOnSSite: recombinationProbabilityOnSSite,
		recombinationProbabilityOnFSite: recombinationProbabilityOnFSite,
	}
}

func calculateAtomFlux(constants config.Constants, temperature float64) float64 {
	const atomMass = 1.67e-27

	v := math.Sqrt((8*1.38*1e-23*temperature)/(math.Pi*constants.Mass*atomMass)) * 1e+2
	atomFlux := 0.25 * v * constants.AgDensity
	return atomFlux
}

func calcR1(constants config.Constants, atomFlux float64) float64 {
	r1 := atomFlux / (constants.FDensity + constants.SDensity)

	return r1
}

func calcR2(constants config.Constants, temperature float64) float64 {
	r2 := math.Exp(-(constants.Edes / (8.31 * temperature))) * constants.Vdes

	return r2
}

func calcR3(constants config.Constants, atomFlux float64) float64 {
	r3 := atomFlux / (constants.FDensity + constants.SDensity)

	return r3
}

func calcR4(constants config.Constants, temperature float64, r3 float64) float64 {
	r4 := math.Exp(-constants.Er/(8.31*temperature)) * r3

	return r4
}

func calcR5(constants config.Constants, temperature float64) float64 {
	r5 := math.Exp(-(constants.Edif / (8.31 * temperature))) * constants.Vdif

	return r5
}

func calcR6(constants config.Constants, temperature float64, r5 float64) float64 {
	r6 := math.Exp(-constants.Er/(8.31*temperature)) * r5

	return r6
}

func calcR7(constants config.Constants, temperature float64, r5 float64) float64 {
	r7 := math.Exp(-constants.Erlh/(8.31*temperature)) * r5

	return r7
}

func calcRecombinationProbabilityOnSSite(constants config.Constants, temperature float64) float64 {
	return math.Exp(-constants.Er / (8.31 * float64(temperature)))
}

func calcRecombinationProbabilityOnFSite(constants config.Constants, temperature float64) float64 {
	return math.Exp(-constants.Erlh / (8.31 * float64(temperature)))
}
