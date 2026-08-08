package converter

import (
	"github.com/stn1slv/md-paste/internal/models"
)

// FlattenTable normalizes a table with merged cells into a perfect grid by repeating content.
func FlattenTable(table models.Table) models.Table {
	if len(table.Rows) == 0 {
		return table
	}

	g := newGrid(len(table.Rows))
	for r, row := range table.Rows {
		col := 0
		for _, sourceCell := range row.Cells {
			col = g.nextFree(r, col)
			col += g.fill(sourceCell, r, col)
		}
	}

	return g.rebuild(table)
}

// grid is a cell grid whose rows widen on demand. The width cannot be derived
// from the source rows up front: a rowspan reaching down from an earlier row
// claims columns in later rows, pushing their own cells to the right, so a row
// can end up wider than its cell count. Growing on placement keeps every cell.
type grid struct {
	cells    [][]models.Cell
	occupied [][]bool
}

func newGrid(rowCount int) *grid {
	return &grid{
		cells:    make([][]models.Cell, rowCount),
		occupied: make([][]bool, rowCount),
	}
}

// ensureWidth grows row r to at least width columns, padding with empty cells
// that carry default spans and alignment rather than zero values.
func (g *grid) ensureWidth(r, width int) {
	for len(g.cells[r]) < width {
		g.cells[r] = append(g.cells[r], models.Cell{
			RowSpan:   1,
			ColSpan:   1,
			Alignment: models.AlignNone,
		})
		g.occupied[r] = append(g.occupied[r], false)
	}
}

// nextFree returns the first column at or after col that is still free in row r.
func (g *grid) nextFree(r, col int) int {
	for col < len(g.occupied[r]) && g.occupied[r][col] {
		col++
	}
	return col
}

// fill writes sourceCell into every position its spans cover and returns the
// number of columns it consumed. A rowspan reaching past the last row is
// truncated; columns are never truncated.
func (g *grid) fill(sourceCell models.Cell, r, col int) int {
	rowSpan := max(sourceCell.RowSpan, 1)
	colSpan := max(sourceCell.ColSpan, 1)

	targetCell := models.Cell{
		Content:   sourceCell.Content,
		Alignment: sourceCell.Alignment,
		RowSpan:   1,
		ColSpan:   1,
	}

	for dr := 0; dr < rowSpan && r+dr < len(g.cells); dr++ {
		g.ensureWidth(r+dr, col+colSpan)
		for dc := 0; dc < colSpan; dc++ {
			g.cells[r+dr][col+dc] = targetCell
			g.occupied[r+dr][col+dc] = true
		}
	}

	return colSpan
}

// rebuild pads every row to the widest one and returns the flattened table.
func (g *grid) rebuild(table models.Table) models.Table {
	maxCols := 0
	for _, row := range g.cells {
		if len(row) > maxCols {
			maxCols = len(row)
		}
	}

	newTable := models.Table{
		HasHeader: table.HasHeader,
		Rows:      make([]models.Row, len(g.cells)),
	}
	for r := range g.cells {
		g.ensureWidth(r, maxCols)
		newTable.Rows[r] = models.Row{
			IsHeader: table.Rows[r].IsHeader,
			Cells:    g.cells[r],
		}
	}
	return newTable
}
