package internal

const (
	Mass      = 14 * 1.67 * 1e-27 // mass of nitrogen atom, kg
	Edes      = 51000             // activation energy of termal desorption, J/mol
	Edif      = 25500             // activation energy of diffusion  Edes/2, J/mol
	Vdes      = 1 * 1e+15         // desorption frequency, 1/s
	Vdif      = 1 * 1e+13         // diffusion frequency? 1/s
	Er        = 14000             // activation energy of recompination  A (Af)+ As -> A2 + Sv + (Fv) , J/mol
	Erlh      = 0                 // activation energy of recompination  Af + Af -> A2 + Fv + Fv , J/mol
	FDensity  = 1.5 * 1e+15       // density of f-sites per cm2
	Fi        = 0.002             // fration of s-sites
	SDensity  = FDensity * Fi     // density of s-sites per cm2
	AgDensity = 1.0 * 1e+15       // density of gas atom per cm2
)
