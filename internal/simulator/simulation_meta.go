package simulator

import (
	"main/internal"
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

func Fill(temperature float64) SimulationMeta {
	atomFlux := calculateAtomFlux(temperature)
	r1 := calcR1(atomFlux)
	r2 := calcR2(temperature)
	r3 := calcR3(atomFlux)
	r4 := calcR4(temperature, r3)
	r5 := calcR5(temperature)
	r6 := calcR6(temperature, r5)
	r7 := calcR7(temperature, r5)

	recombinationProbabilityOnSSite := calcRecombinationProbabilityOnSSite(temperature)
	recombinationProbabilityOnFSite := calcRecombinationProbabilityOnFSite(temperature)

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

func calculateAtomFlux(temperature float64) float64 {
	v := math.Sqrt((8*1.38*10e-23*temperature)/(math.Pi*internal.Mass)) * 10e+2
	atomFlux := 0.25 * v * 10e+15
	return atomFlux
}

func calcR1(atomFlux float64) float64 {
	r1 := atomFlux / (internal.FDensity + internal.SDensity)

	return r1
}

func calcR2(temperature float64) float64 {
	r2 := math.Exp(-(internal.Edes / (8.31 * temperature))) * internal.Vdes

	return r2
}

func calcR3(atomFlux float64) float64 {
	r3 := atomFlux / (internal.FDensity + internal.SDensity)

	return r3
}

func calcR4(temperature float64, r3 float64) float64 {
	r4 := math.Exp(-internal.Er/(8.31*temperature)) * r3

	return r4
}

func calcR5(temperature float64) float64 {
	r5 := math.Exp(-(internal.Edif / (8.31 * temperature))) * internal.Vdif

	return r5
}

func calcR6(temperature float64, r5 float64) float64 {
	r6 := math.Exp(-internal.Er/(8.31*temperature)) * r5

	return r6
}

func calcR7(temperature float64, r5 float64) float64 {
	r7 := math.Exp(-internal.Erlh/(8.31*temperature)) * r5

	return r7
}

func calcRecombinationProbabilityOnSSite(temperature float64) float64 {
	return math.Exp(-internal.Er / (8.31 * float64(temperature)))
}

func calcRecombinationProbabilityOnFSite(temperature float64) float64 {
	return math.Exp(-internal.Erlh / (8.31 * float64(temperature)))
}
