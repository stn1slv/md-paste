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
	// If the source has no header, we treat the first row as the header.
	headerRow := table.Rows[0]
	renderRow(&sb, headerRow)
	sb.WriteString("\n")

	// Render separator row
	renderSeparator(&sb, headerRow)
	sb.WriteString("\n")

	// Render remaining rows
	for i := 1; i < len(table.Rows); i++ {
		renderRow(&sb, table.Rows[i])
		if i < len(table.Rows)-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func renderRow(sb *strings.Builder, row models.Row) {
	sb.WriteString("|")
	for _, cell := range row.Cells {
		content := escapePipe(cell.Content)
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

func escapePipe(content string) string {
	return strings.ReplaceAll(content, "|", "\\|")
}
