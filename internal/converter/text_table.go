package converter

import (
	"regexp"
	"strings"

	"github.com/stn1slv/md-paste/internal/models"
)

var colSeparator = regexp.MustCompile(`\s{2,}|\t`)

// ExtractTableFromText attempts to reconstruct a table from plain text using layout heuristics.
func ExtractTableFromText(text string) (models.Table, bool) {
	// Only surrounding blank lines are stripped. Leading whitespace on the first
	// line is significant to dedent below, so TrimSpace must not eat it.
	lines := strings.Split(strings.Trim(text, "\r\n"), "\n")
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
	mostCommonWidthCount := 0
	for _, n := range countByWidth {
		if n > mostCommonWidthCount {
			mostCommonWidthCount = n
		}
	}
	return mostCommonWidthCount*2 >= multiColRows
}

func parseTextToRows(lines []string) models.Table {
	var table models.Table
	for _, line := range dedent(lines) {
		if strings.TrimSpace(line) == "" {
			continue
		}

		parts := colSeparator.Split(line, -1)

		cleanParts := make([]string, 0, len(parts))
		for _, p := range parts {
			cleanParts = append(cleanParts, strings.TrimSpace(p))
		}

		// Remove trailing empty parts introduced by trailing whitespace. Leading
		// ones are kept: they are a genuinely empty first column.
		for len(cleanParts) > 0 && cleanParts[len(cleanParts)-1] == "" {
			cleanParts = cleanParts[:len(cleanParts)-1]
		}

		if len(cleanParts) > 0 {
			table.Rows = append(table.Rows, buildTextRow(cleanParts))
		}
	}
	return table
}

// dedent strips the longest leading-whitespace prefix shared by every non-blank
// line. Uniform indentation would otherwise match the column separator and give
// every row a spurious empty first column, while trimming each line on its own
// would discard the genuinely empty first column of a continuation row (a row
// whose first cell is blank is indented past where that cell would start).
func dedent(lines []string) []string {
	prefix := ""
	found := false
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		indent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
		if !found {
			prefix, found = indent, true
			continue
		}
		if prefix = commonPrefix(prefix, indent); prefix == "" {
			return lines
		}
	}
	if prefix == "" {
		return lines
	}

	out := make([]string, len(lines))
	for i, line := range lines {
		out[i] = strings.TrimPrefix(line, prefix)
	}
	return out
}

func commonPrefix(a, b string) string {
	n := min(len(a), len(b))
	i := 0
	for i < n && a[i] == b[i] {
		i++
	}
	return a[:i]
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
