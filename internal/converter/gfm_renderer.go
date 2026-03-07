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

	// GFM requires a header row and a separator row.
	// We treat the first row (index 0) as the header to preserve original row order,
	// ignoring IsHeader if it's not the first row, to avoid reordering.
	headerRow := table.Rows[0]
	renderRow(&sb, headerRow)
	sb.WriteString("\n")

	// Render separator row based on header alignment
	renderSeparator(&sb, headerRow)

	// Render all other rows
	for i := 1; i < len(table.Rows); i++ {
		sb.WriteString("\n")
		renderRow(&sb, table.Rows[i])
	}

	return sb.String()
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
	// We replace them with spaces to preserve structure, but avoid
	// collapsing other whitespace that may be meaningful in Markdown
	// (e.g., inside inline code spans).
	content = strings.ReplaceAll(content, "\r\n", " ")
	content = strings.ReplaceAll(content, "\n", " ")
	content = strings.ReplaceAll(content, "\r", " ")
	// Trim leading and trailing spaces that may have been introduced
	// by newline replacement, while preserving internal spacing.
	content = strings.TrimSpace(content)
	// Escape pipes
	return strings.ReplaceAll(content, "|", "\\|")
}
