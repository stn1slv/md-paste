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

func isTabular(table models.Table) bool {
	if len(table.Rows) < 2 {
		return false
	}

	multiColRows := 0
	listLikeRows := 0
	for _, row := range table.Rows {
		if len(row.Cells) > 1 {
			multiColRows++
			if len(row.Cells) == 2 && listMarker.MatchString(row.Cells[0].Content) {
				listLikeRows++
			}
		}
	}

	// Heuristic 1: If it's just a 2-column list, don't treat it as a table.
	if listLikeRows > 0 && listLikeRows*2 >= multiColRows {
		return false
	}

	// Heuristic 2: Majority of rows should have multiple columns to be considered a table.
	// This filters out regular text blocks.
	return multiColRows*2 >= len(table.Rows)
}

func parseTextToRows(lines []string) models.Table {
	var table models.Table
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Split by the separator (2+ spaces or tab)
		parts := colSeparator.Split(line, -1)

		cleanParts := make([]string, 0, len(parts))
		for _, p := range parts {
			cleanParts = append(cleanParts, strings.TrimSpace(p))
		}

		// Remove trailing empty parts introduced by trailing whitespace
		for len(cleanParts) > 0 && cleanParts[len(cleanParts)-1] == "" {
			cleanParts = cleanParts[:len(cleanParts)-1]
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
