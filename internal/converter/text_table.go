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
	if len(table.Rows) < 2 {
		return models.Table{}, false
	}

	normalizeColumnCounts(&table)
	return table, true
}

func parseTextToRows(lines []string) models.Table {
	var table models.Table
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := colSeparator.Split(line, -1)
		cleanParts := cleanRowParts(parts)
		if len(cleanParts) > 1 {
			table.Rows = append(table.Rows, buildTextRow(cleanParts))
		}
	}
	return table
}

func cleanRowParts(parts []string) []string {
	var cleanParts []string
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			cleanParts = append(cleanParts, s)
		}
	}
	return cleanParts
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
		// No truncation needed if we use maxCols
	}
}

func padRow(row *models.Row, targetCount int) {
	for j := len(row.Cells); j < targetCount; j++ {
		row.Cells = append(row.Cells, models.Cell{Content: "", RowSpan: 1, ColSpan: 1})
	}
}
