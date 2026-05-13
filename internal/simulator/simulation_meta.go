package simulator

import (
	"main/configs"
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

func Fill(element configs.Element, constants configs.Constants, temperature float64) SimulationMeta {
	atomFlux := calculateAtomFlux(element, temperature)
	r1 := calcR1(constants, atomFlux)
	r2 := calcR2(element, temperature)
	r3 := calcR3(constants, atomFlux)
	r4 := calcR4(element, temperature, r3)
	r5 := calcR5(element, temperature)
	r6 := calcR6(element, temperature, r5)
	r7 := calcR7(element, temperature, r5)

	recombinationProbabilityOnSSite := calcRecombinationProbabilityOnSSite(element, temperature)
	recombinationProbabilityOnFSite := calcRecombinationProbabilityOnFSite(element, temperature)

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

func calculateAtomFlux(element configs.Element, temperature float64) float64 {
	const atomMass = 1.66035e-27

	v := math.Sqrt((8*1.38*1e-23*temperature)/(math.Pi*element.Mass*atomMass)) * 1e+2
	atomFlux := 0.25 * v * element.AgDensity
	return atomFlux
}

func calcR1(constants configs.Constants, atomFlux float64) float64 {
	r1 := atomFlux / (constants.FDensity + constants.SDensity)
	return r1
}

func calcR2(element configs.Element, temperature float64) float64 {
	r2 := math.Exp(-(element.Edes / (8.31 * temperature))) * element.Vdes
	return r2
}

func calcR3(constants configs.Constants, atomFlux float64) float64 {
	r3 := atomFlux / (constants.FDensity + constants.SDensity)
	return r3
}

func calcR4(element configs.Element, temperature float64, r3 float64) float64 {
	r4 := math.Exp(-element.Er/(8.31*temperature)) * r3
	return r4
}

func calcR5(element configs.Element, temperature float64) float64 {
	r5 := math.Exp(-(element.Edif / (8.31 * temperature))) * element.Vdif
	return r5
}

func calcR6(element configs.Element, temperature float64, r5 float64) float64 {
	r6 := math.Exp(-element.Er/(8.31*temperature)) * r5
	return r6
}

func calcR7(element configs.Element, temperature float64, r5 float64) float64 {
	r7 := math.Exp(-element.Erlh/(8.31*temperature)) * r5
	return r7
}

// TODO: для разных элементов добавить
func calcRecombinationProbabilityOnSSite(element configs.Element, temperature float64) float64 {
	return math.Exp(-element.Er / (8.31 * float64(temperature)))
}

func calcRecombinationProbabilityOnFSite(element configs.Element, temperature float64) float64 {
	return math.Exp(-element.Erlh / (8.31 * float64(temperature)))
}
