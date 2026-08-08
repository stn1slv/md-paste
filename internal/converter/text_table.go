package converter

import (
	"regexp"
	"strings"

	"github.com/stn1slv/md-paste/internal/models"
)

var colSeparator = regexp.MustCompile(`\s{2,}|\t`)

// ExtractTableFromText attempts to reconstruct a table from plain text using layout heuristics.
func ExtractTableFromText(text string) (models.Table, bool) {
	lines := strings.Split(strings.TrimSpace(text), "\n")
	if len(lines) < 2 {
		return models.Table{}, false
	}

	table := parseTextToRows(lines)

	if !isTabular(table) {
		return models.Table{}, false
	}

	normalizeColumnCounts(&table)
	return table, true
}

var listMarker = regexp.MustCompile(`^(\d+[\.\)]|[-*•+])\s*$`)

// minMultiColRows is the smallest number of multi-column rows that can look like
// a table. A single such row inside a block of prose is far more likely to be a
// sentence containing a double space than a table.
const minMultiColRows = 2

func isTabular(table models.Table) bool {
	if len(table.Rows) < 2 {
		return false
	}

	multiColRows := 0
	listLikeRows := 0
	// countByWidth maps a column count to how many rows have exactly that many.
	countByWidth := make(map[int]int)
	for _, row := range table.Rows {
		if len(row.Cells) > 1 {
			multiColRows++
			countByWidth[len(row.Cells)]++
			if len(row.Cells) == 2 && listMarker.MatchString(row.Cells[0].Content) {
				listLikeRows++
			}
		}
	}

	// Heuristic 1: If it's just a 2-column list, don't treat it as a table.
	if listLikeRows > 0 && listLikeRows*2 >= multiColRows {
		return false
	}

	// Heuristic 2: One multi-column row is not evidence of a table.
	if multiColRows < minMultiColRows {
		return false
	}

	// Heuristic 3: At least two thirds of the rows must have multiple columns.
	// This filters out prose blocks where only some lines happen to contain a
	// double space (for example after a sentence-ending period).
	if multiColRows*3 < len(table.Rows)*2 {
		return false
	}

	// Heuristic 4: Real tables are close to rectangular. Require the most common
	// column count to cover at least half of the multi-column rows, so text whose
	// rows disagree wildly on their width is rejected.
	widest := 0
	for _, n := range countByWidth {
		if n > widest {
			widest = n
		}
	}
	return widest*2 >= multiColRows
}

func parseTextToRows(lines []string) models.Table {
	var table models.Table
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Split the trimmed line: leading indentation would otherwise match the
		// separator and produce a spurious empty first column.
		parts := colSeparator.Split(trimmed, -1)

		cleanParts := make([]string, 0, len(parts))
		for _, p := range parts {
			cleanParts = append(cleanParts, strings.TrimSpace(p))
		}

		if len(cleanParts) > 0 {
			table.Rows = append(table.Rows, buildTextRow(cleanParts))
		}
	}
	return table
}

func buildTextRow(parts []string) models.Row {
	var row models.Row
	for _, p := range parts {
		row.Cells = append(row.Cells, models.Cell{Content: p, RowSpan: 1, ColSpan: 1})
	}
	return row
}

func normalizeColumnCounts(table *models.Table) {
	if len(table.Rows) == 0 {
		return
	}

	// Compute max column count across all rows to avoid dropping data
	maxCols := 0
	for _, row := range table.Rows {
		if len(row.Cells) > maxCols {
			maxCols = len(row.Cells)
		}
	}

	for i := range table.Rows {
		if len(table.Rows[i].Cells) < maxCols {
			padRow(&table.Rows[i], maxCols)
		}
	}
}

func padRow(row *models.Row, targetCount int) {
	for j := len(row.Cells); j < targetCount; j++ {
		row.Cells = append(row.Cells, models.Cell{Content: "", RowSpan: 1, ColSpan: 1})
	}
}
