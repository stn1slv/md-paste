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

	// GFM requires the header to be the first row, so rows are always emitted in
	// their original order and row 0 becomes the header. When a later row is the
	// one marked as the header (an HTML table whose first <th> row is not the
	// first row), its column alignment still drives the separator, but no row is
	// moved: reordering would silently scramble the data.
	headerRow := table.Rows[0]
	alignRow := headerRow
	for _, row := range table.Rows {
		if row.IsHeader {
			alignRow = row
			break
		}
	}

	renderRow(&sb, headerRow)
	sb.WriteString("\n")
	renderSeparator(&sb, headerRow, alignRow)

	for _, row := range table.Rows[1:] {
		sb.WriteString("\n")
		renderRow(&sb, row)
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

// renderSeparator emits one separator column per header cell. Each column's
// alignment comes from alignRow, which may be a different row than the one
// rendered as the header; columns missing there fall back to no alignment.
func renderSeparator(sb *strings.Builder, headerRow, alignRow models.Row) {
	sb.WriteString("|")
	for i := range headerRow.Cells {
		alignment := models.AlignNone
		if i < len(alignRow.Cells) {
			alignment = alignRow.Cells[i].Alignment
		}
		sb.WriteString(" ")
		//nolint:exhaustive // Default handles AlignNone and any future alignments
		switch alignment {
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
