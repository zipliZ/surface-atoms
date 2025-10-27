package simulator

import (
	"crypto/rand"
	"main/configs"
	"main/internal/generators"
	"main/internal/random"
	"math/big"
)

type SurfaceAtomsController struct {
	AtomsOnSurface  map[int]Atom
	AtomsOnFCenters AtomsOnCenters
	AtomsOnSCenters AtomsOnCenters
	MatrixLimitX    int
	MatrixLimitY    int
	matrix          *Matrix
	IdGenerator     *generators.IdGenerator
}

func NewSurfaceAtomsController(matrixLimitX int, matrixLimitY int, matrix *Matrix, elements []configs.Element) *SurfaceAtomsController {
	atomsOnSurface := make(map[int]Atom)
	atomsOnFCenters := make(map[string]*random.Map[int, Atom])
	atomsOnSCenters := make(map[string]*random.Map[int, Atom])
	for _, element := range elements {
		atomsOnFCenters[element.Name] = random.NewRandMap[int, Atom]()
		atomsOnSCenters[element.Name] = random.NewRandMap[int, Atom]()
	}

	return &SurfaceAtomsController{
		AtomsOnSurface:  atomsOnSurface,
		MatrixLimitX:    matrixLimitX,
		MatrixLimitY:    matrixLimitY,
		matrix:          matrix,
		AtomsOnFCenters: atomsOnFCenters,
		AtomsOnSCenters: atomsOnSCenters,
		IdGenerator:     generators.NewIdGenerator(),
	}
}

type AtomsOnCenters map[string]*random.Map[int, Atom]

func (a AtomsOnCenters) Len() int {
	total := 0
	for _, atoms := range a {
		total += atoms.Len()
	}
	return total
}

type Atom struct {
	Id             int
	X              uint32
	Y              uint32
	OccupiedCentre rune
	ElementName    string
}

func (a *Atom) ChangePosition(x uint32, y uint32, center rune) {
	a.X = x
	a.Y = y
	a.OccupiedCentre = center
}

func (s *SurfaceAtomsController) AddAtomOnSurface(atom Atom) {
	atom.Id = s.IdGenerator.Generate()
	s.AtomsOnSurface[atom.Id] = atom

	switch atom.OccupiedCentre {
	case 'F':
		s.AtomsOnFCenters[atom.ElementName].Add(atom.Id, atom)
	case 'S':
		s.AtomsOnSCenters[atom.ElementName].Add(atom.Id, atom)
	}

	s.matrix.SetAtomOnCell(atom.X, atom.Y, atom.Id)
}

func (s *SurfaceAtomsController) GetNextAtomCoordinates(atomId int) (x, y uint32) {
	movement := map[int32]struct {
		x int32
		y int32
	}{
		1: {-1, 0},
		2: {0, 1},
		3: {1, 0},
		4: {0, -1},
	}
	var (
		nextX, nextY uint32 = 0, 0
	)

	atom := s.AtomsOnSurface[atomId]

	positionNotFound := true
	for positionNotFound {
		randomDirection, _ := rand.Int(rand.Reader, big.NewInt(4))
		randomInt := int32(randomDirection.Int64() + 1)
		possibleX := int32(atom.X) + movement[randomInt].x
		possibleY := int32(atom.Y) + movement[randomInt].y

		if (0 <= possibleX && possibleX < int32(s.MatrixLimitX)) &&
			(0 <= possibleY && possibleY < int32(s.MatrixLimitY)) {
			nextX, nextY = uint32(possibleX), uint32(possibleY)
			positionNotFound = false
		}
	}

	return nextX, nextY
}

func (s *SurfaceAtomsController) RemoveAtomFromSurface(atomId int) {
	atom := s.AtomsOnSurface[atomId]

	switch atom.OccupiedCentre {
	case 'F':
		s.AtomsOnFCenters[atom.ElementName].Remove(atomId)
	case 'S':
		s.AtomsOnSCenters[atom.ElementName].Remove(atomId)
	}

	delete(s.AtomsOnSurface, atomId)
	s.matrix.ClearCell(atom.X, atom.Y)
}

func (s *SurfaceAtomsController) MoveAtom(atom Atom, nextCell CellData) {
	switch atom.OccupiedCentre {
	case 'F':
		s.AtomsOnFCenters[atom.ElementName].Remove(atom.Id)
	case 'S':
		s.AtomsOnSCenters[atom.ElementName].Remove(atom.Id)
	}

	s.matrix.ClearCell(atom.X, atom.Y)
	atom.ChangePosition(nextCell.X, nextCell.Y, nextCell.Center)
	s.AtomsOnSurface[atom.Id] = atom
	s.matrix.SetAtomOnCell(nextCell.X, nextCell.Y, atom.Id)

	switch nextCell.Center {
	case 'F':
		s.AtomsOnFCenters[atom.ElementName].Add(atom.Id, atom)
	case 'S':
		s.AtomsOnSCenters[atom.ElementName].Add(atom.Id, atom)
	}
}
