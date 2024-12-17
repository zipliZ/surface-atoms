package internal

import (
	"crypto/rand"
	"math/big"
)

type Matrix struct {
	NumOfSSites         int
	NumOfFSites         int
	FreeCellsOfFCenters *RandMap[int, CellData]
	FreeCellsOfSCenters *RandMap[int, CellData]
	cells               [][]CellData
}

// CellData представляет данные об отдельной ячейке в матрице.
// Поля:
// - Id: уникальный идентификатор ячейки.
// - X, Y: координаты ячейки в матрице.
// - Center: тип центра ('S' или 'F'), к которому относится ячейка.
// - IsFree: флаг, указывающий, свободна ли ячейка.
// - AtomId: идентификатор атома, находящегося в ячейке (если есть).
type CellData struct {
	Id     int
	X      int
	Y      int
	Center rune
	IsFree bool
	AtomId int
}

func NewMatrix() *Matrix {
	return &Matrix{
		cells:               [][]CellData{},
		FreeCellsOfSCenters: NewRandMap[int, CellData](),
		FreeCellsOfFCenters: NewRandMap[int, CellData](),
	}
}

// Init - инициализирует матрицу заданным размером.
// Заполняет матрицу данными, вычисляет количество S- и F-центров.
func (m *Matrix) Init(x, y int) {
	m.cells = make([][]CellData, x)

	m.NumOfSSites = int(float64(x) * float64(y) * Fi)
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

// SetAtomOnCell - ставит атом на ячейку (x,y) с id atomId.
// Если ячейка была свободной, то она перестает быть свободной.
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

// ClearCell - очищает ячейку (x,y) - удаляет атом, если он там был.
// Если ячейка была свободной, то она продолжает быть свободной.
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

// GetCellInfo - возвращает информацию о ячейке (x,y)
func (m *Matrix) GetCellInfo(x, y int) CellData {
	return m.cells[y][x]
}

// CountFreeCellsOfFCenters - возвращает количество свободных F-центров
func (m *Matrix) CountFreeCellsOfFCenters() int {
	return m.FreeCellsOfFCenters.Len()
}

// CountFreeCellsOFSCenters - возвращает количество свободных S-центров
func (m *Matrix) CountFreeCellsOFSCenters() int {
	return m.FreeCellsOfSCenters.Len()
}
