package converter

import (
	"github.com/stn1slv/md-paste/internal/models"
)

// FlattenTable normalizes a table with merged cells into a perfect grid by repeating content.
func FlattenTable(table models.Table) models.Table {
	if len(table.Rows) == 0 {
		return table
	}

	maxCols := calculateMaxColumns(table)
	rowCount := len(table.Rows)

	grid := make([][]models.Cell, rowCount)
	occupied := make([][]bool, rowCount)
	for i := 0; i < rowCount; i++ {
		grid[i] = make([]models.Cell, maxCols)
		occupied[i] = make([]bool, maxCols)
	}

	populateGrid(table, grid, occupied, rowCount, maxCols)

	return rebuildTable(table, grid, rowCount)
}

func calculateMaxColumns(table models.Table) int {
	maxCols := 0
	for _, row := range table.Rows {
		width := 0
		for _, cell := range row.Cells {
			if cell.ColSpan > 1 {
				width += cell.ColSpan
			} else {
				width++
			}
		}
		if width > maxCols {
			maxCols = width
		}
	}
	return maxCols
}

func populateGrid(table models.Table, grid [][]models.Cell, occupied [][]bool, rowCount, maxCols int) {
	for r, row := range table.Rows {
		col := 0
		for _, sourceCell := range row.Cells {
			for col < maxCols && occupied[r][col] {
				col++
			}
			if col >= maxCols {
				break
			}
			col += fillSpan(sourceCell, grid, occupied, r, col, rowCount, maxCols)
		}
	}
}

func fillSpan(sourceCell models.Cell, grid [][]models.Cell, occupied [][]bool, r, col, rowCount, maxCols int) int {
	rowSpan := sourceCell.RowSpan
	if rowSpan < 1 {
		rowSpan = 1
	}
	colSpan := sourceCell.ColSpan
	if colSpan < 1 {
		colSpan = 1
	}

	targetCell := models.Cell{
		Content:   sourceCell.Content,
		Alignment: sourceCell.Alignment,
		RowSpan:   1,
		ColSpan:   1,
	}

	for dr := 0; dr < rowSpan; dr++ {
		for dc := 0; dc < colSpan; dc++ {
			targetR, targetC := r+dr, col+dc
			if targetR < rowCount && targetC < maxCols {
				grid[targetR][targetC] = targetCell
				occupied[targetR][targetC] = true
			}
		}
	}
	return colSpan
}

func rebuildTable(table models.Table, grid [][]models.Cell, rowCount int) models.Table {
	newTable := models.Table{
		HasHeader: table.HasHeader,
		Rows:      make([]models.Row, rowCount),
	}
	for r := 0; r < rowCount; r++ {
		newTable.Rows[r] = models.Row{
			IsHeader: table.Rows[r].IsHeader,
			Cells:    grid[r],
		}
	}
	return newTable
}
