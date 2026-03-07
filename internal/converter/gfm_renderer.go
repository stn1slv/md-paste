package converter

import (
	"strings"

	"github.com/stn1slv/md-paste/internal/models"
)

// RenderTable converts a Table model into a GFM pipe table string.
func RenderTable(table models.Table) string {
	if len(table.Rows) == 0 {
		return ""
	}

	var sb strings.Builder

	// Find the header row. Use the first row marked as header, or fallback to index 0.
	headerIdx := findHeaderIndex(table)
	headerRow := table.Rows[headerIdx]

	// Render header row
	renderRow(&sb, headerRow)
	sb.WriteString("\n")

	// Render separator row based on header alignment
	renderSeparator(&sb, headerRow)

	// Render all other rows
	for i, row := range table.Rows {
		if i == headerIdx {
			continue
		}
		sb.WriteString("\n")
		renderRow(&sb, row)
	}

	return sb.String()
}

func findHeaderIndex(table models.Table) int {
	for i, row := range table.Rows {
		if row.IsHeader {
			return i
		}
	}
	return 0
}

func renderRow(sb *strings.Builder, row models.Row) {
	sb.WriteString("|")
	for _, cell := range row.Cells {
		content := sanitizeCellContent(cell.Content)
		sb.WriteString(" ")
		sb.WriteString(content)
		sb.WriteString(" |")
	}
}

func renderSeparator(sb *strings.Builder, headerRow models.Row) {
	sb.WriteString("|")
	for _, cell := range headerRow.Cells {
		sb.WriteString(" ")
		switch cell.Alignment {
		case models.AlignLeft:
			sb.WriteString(":---")
		case models.AlignCenter:
			sb.WriteString(":---:")
		case models.AlignRight:
			sb.WriteString("---:")
		default:
			sb.WriteString("---")
		}
		sb.WriteString(" |")
	}
}

func sanitizeCellContent(content string) string {
	// GFM tables do not support newlines within cells.
	// We replace them with spaces to preserve structure.
	content = strings.ReplaceAll(content, "\n", " ")
	// Collapse multiple spaces that might have been introduced
	content = strings.Join(strings.Fields(content), " ")
	// Escape pipes
	return strings.ReplaceAll(content, "|", "\\|")
}
