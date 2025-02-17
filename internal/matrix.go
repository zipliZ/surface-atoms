package internal

import (
	"crypto/rand"
	"main/internal/config"
	"math/big"
)

type Matrix struct {
	NumOfSSites         int
	NumOfFSites         int
	FreeCellsOfFCenters *RandMap[int, CellData]
	FreeCellsOfSCenters *RandMap[int, CellData]
	cells               [][]CellData
	consts              config.Constants
}

// CellData represents the data of an individual cell in the matrix.
// Fields:
// - Id: unique identifier of the cell.
// - X, Y: coordinates of the cell in the matrix.
// - Center: type of center ('S' or 'F') the cell belongs to.
// - IsFree: flag indicating whether the cell is free.
// - AtomId: identifier of the atom present in the cell (if any).
type CellData struct {
	Id     int
	X      int
	Y      int
	Center rune
	IsFree bool
	AtomId int
}

func NewMatrix(consts config.Constants) *Matrix {
	return &Matrix{
		cells:               [][]CellData{},
		FreeCellsOfSCenters: NewRandMap[int, CellData](),
		FreeCellsOfFCenters: NewRandMap[int, CellData](),
		consts:              consts,
	}
}

// Init initializes the matrix with the given size.
// It fills the matrix with data and calculates the number of S- and F-centers.
func (m *Matrix) Init(x, y int) {
	m.cells = make([][]CellData, x)

	m.NumOfSSites = int(float64(x) * float64(y) * m.consts.Fi)
	m.NumOfFSites = x*y - m.NumOfSSites

	for i := range m.cells {
		m.cells[i] = make([]CellData, y)
		for j := range m.cells[i] {
			m.cells[i][j].Id = i*x + j + 1
			m.cells[i][j].X = j
			m.cells[i][j].Y = i
			m.cells[i][j].IsFree = true
			m.cells[i][j].Center = 'F'
			m.FreeCellsOfFCenters.Add(m.cells[i][j].Id, m.cells[i][j])
		}
	}

	for range m.NumOfSSites {
		for {
			randomI, _ := rand.Int(rand.Reader, big.NewInt(int64(x-1)))
			randomJ, _ := rand.Int(rand.Reader, big.NewInt(int64(y-1)))

			i, j := int(randomI.Int64()), int(randomJ.Int64())
			if m.cells[j][i].Center != 'S' {
				m.cells[j][i].Center = 'S'
				m.FreeCellsOfSCenters.Add(m.cells[j][i].Id, m.cells[j][i])
				m.FreeCellsOfFCenters.Remove(m.cells[j][i].Id)
				break
			}
		}
	}
}

// SetAtomOnCell places an atom on the cell (x, y) with the given atomId.
// If the cell was free, it is no longer considered free.
func (m *Matrix) SetAtomOnCell(x, y int, atomId int) {
	m.cells[y][x].IsFree = false
	m.cells[y][x].AtomId = atomId

	switch m.cells[y][x].Center {
	case 'S':
		m.FreeCellsOfSCenters.Remove(m.cells[y][x].Id)
	case 'F':
		m.FreeCellsOfFCenters.Remove(m.cells[y][x].Id)
	}
}

// ClearCell clears the cell (x, y), removing the atom if present.
// If the cell was already free, it remains free.
func (m *Matrix) ClearCell(x, y int) {
	m.cells[y][x].IsFree = true
	m.cells[y][x].AtomId = 0

	switch m.cells[y][x].Center {
	case 'S':
		m.FreeCellsOfSCenters.Add(m.cells[y][x].Id, m.cells[y][x])
	case 'F':
		m.FreeCellsOfFCenters.Add(m.cells[y][x].Id, m.cells[y][x])
	}
}

// GetCellInfo returns information about the cell at (x, y).
func (m *Matrix) GetCellInfo(x, y int) CellData {
	return m.cells[y][x]
}

// CountFreeCellsOfFCenters returns the number of free F-centers.
func (m *Matrix) CountFreeCellsOfFCenters() int {
	return m.FreeCellsOfFCenters.Len()
}

// CountFreeCellsOfSCenters returns the number of free S-centers.
func (m *Matrix) CountFreeCellsOfSCenters() int {
	return m.FreeCellsOfSCenters.Len()
}
