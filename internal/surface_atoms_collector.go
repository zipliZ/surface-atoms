package internal

import (
	"crypto/rand"
	"math/big"
)

type SurfaceAtomsController struct {
	AtomsOnSurface  map[int]Atom
	AtomsOnFCenters *RandMap[int, Atom]
	AtomsOnSCenters *RandMap[int, Atom]
	MatrixLimitX    int
	MatrixLimitY    int
	matrix          *Matrix
	*IdGenerator
}

func NewSurfaceAtomsController(matrixLimitX int, matrixLimitY int, matrix *Matrix) *SurfaceAtomsController {
	atomsOnSurface := make(map[int]Atom)
	atomsOnFCenters := NewRandMap[int, Atom]()
	atomsOnCCenters := NewRandMap[int, Atom]()

	return &SurfaceAtomsController{
		AtomsOnSurface:  atomsOnSurface,
		MatrixLimitX:    matrixLimitX,
		MatrixLimitY:    matrixLimitY,
		matrix:          matrix,
		AtomsOnFCenters: atomsOnFCenters,
		AtomsOnSCenters: atomsOnCCenters,
		IdGenerator:     NewIdGenerator(),
	}
}

type Atom struct {
	Id             int
	X              int
	Y              int
	OccupiedCentre rune
}

func (a *Atom) ChangePosition(x int, y int, center rune) {
	a.X = x
	a.Y = y
	a.OccupiedCentre = center
}

func (s *SurfaceAtomsController) AddAtomOnSurface(atom Atom) {
	atom.Id = s.GetId()
	s.AtomsOnSurface[atom.Id] = atom

	switch atom.OccupiedCentre {
	case 'F':
		s.AtomsOnFCenters.Add(atom.Id, atom)
	case 'S':
		s.AtomsOnSCenters.Add(atom.Id, atom)
	}

	s.matrix.SetAtomOnCell(atom.X, atom.Y, atom.Id)
}

func (s *SurfaceAtomsController) GetNextAtomCoordinates(atomId int) (x, y int) {
	movement := map[int64]struct {
		x int
		y int
	}{
		1: {-1, 0},
		2: {0, 1},
		3: {1, 0},
		4: {0, -1},
	}
	nextX, nextY := 0, 0
	atom := s.AtomsOnSurface[atomId]

	positionNotFount := true
	for positionNotFount {
		randomDirection, _ := rand.Int(rand.Reader, big.NewInt(4))
		randomInt := randomDirection.Int64() + 1
		possibleX, possibleY := atom.X+movement[randomInt].x, atom.Y+movement[randomInt].y

		if (0 <= possibleX && possibleX < s.MatrixLimitX) &&
			(0 <= possibleY && possibleY < s.MatrixLimitY) {
			nextX, nextY = possibleX, possibleY
			positionNotFount = false
		}
	}

	return nextX, nextY
}

func (s *SurfaceAtomsController) RemoveAtomFromSurface(atomId int) {
	atom := s.AtomsOnSurface[atomId]

	switch atom.OccupiedCentre {
	case 'F':
		s.AtomsOnFCenters.Remove(atomId)
	case 'S':
		s.AtomsOnSCenters.Remove(atomId)
	}

	delete(s.AtomsOnSurface, atomId)
	s.matrix.ClearCell(atom.X, atom.Y)
}

func (s *SurfaceAtomsController) MoveAtom(atom Atom, nextCell CellData) {
	switch atom.OccupiedCentre {
	case 'F':
		s.AtomsOnFCenters.Remove(atom.Id)
	case 'S':
		s.AtomsOnSCenters.Remove(atom.Id)
	}

	s.matrix.ClearCell(atom.X, atom.Y)
	atom.ChangePosition(nextCell.X, nextCell.Y, nextCell.Center)
	s.AtomsOnSurface[atom.Id] = atom
	s.matrix.SetAtomOnCell(nextCell.X, nextCell.Y, atom.Id)

	switch nextCell.Center {
	case 'F':
		s.AtomsOnFCenters.Add(atom.Id, atom)
	case 'S':
		s.AtomsOnSCenters.Add(atom.Id, atom)
	}
}
