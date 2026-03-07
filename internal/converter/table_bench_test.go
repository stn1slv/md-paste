package converter

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stn1slv/md-paste/internal/models"
)

func BenchmarkRenderTable(b *testing.B) {
	table := models.Table{
		Rows: make([]models.Row, 10),
	}
	for i := 0; i < 10; i++ {
		row := models.Row{Cells: make([]models.Cell, 10)}
		for j := 0; j < 10; j++ {
			row.Cells[j] = models.Cell{Content: fmt.Sprintf("Cell %d-%d", i, j), RowSpan: 1, ColSpan: 1}
		}
		table.Rows[i] = row
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = RenderTable(table)
	}
}

func BenchmarkExtractTableFromHTML(b *testing.B) {
	var sb strings.Builder
	sb.WriteString("<table>")
	for i := 0; i < 10; i++ {
		sb.WriteString("<tr>")
		for j := 0; j < 10; j++ {
			sb.WriteString(fmt.Sprintf("<td>Data %d-%d</td>", i, j))
		}
		sb.WriteString("</tr>")
	}
	sb.WriteString("</table>")
	html := sb.String()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ExtractTableFromHTML(html)
	}
}
