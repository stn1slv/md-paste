package converter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stn1slv/md-paste/internal/models"
)

func TestExtractTableFromHTML(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected models.Table
		found    bool
	}{
		{
			name: "simple table",
			html: `<table><tr><th>H1</th><th>H2</th></tr><tr><td>D1</td><td>D2</td></tr></table>`,
			expected: models.Table{
				HasHeader: true,
				Rows: []models.Row{
					{
						IsHeader: true,
						Cells: []models.Cell{
							{Content: "H1", Alignment: models.AlignNone, RowSpan: 1, ColSpan: 1},
							{Content: "H2", Alignment: models.AlignNone, RowSpan: 1, ColSpan: 1},
						},
					},
					{
						Cells: []models.Cell{
							{Content: "D1", Alignment: models.AlignNone, RowSpan: 1, ColSpan: 1},
							{Content: "D2", Alignment: models.AlignNone, RowSpan: 1, ColSpan: 1},
						},
					},
				},
			},
			found: true,
		},
		{
			name: "table with alignment",
			html: `<table>
				<tr>
					<th align="left">Left</th>
					<th style="text-align: center">Center</th>
					<th style="text-align: right;">Right</th>
				</tr>
				<tr>
					<td>1</td>
					<td>2</td>
					<td>3</td>
				</tr>
			</table>`,
			expected: models.Table{
				HasHeader: true,
				Rows: []models.Row{
					{
						IsHeader: true,
						Cells: []models.Cell{
							{Content: "Left", Alignment: models.AlignLeft, RowSpan: 1, ColSpan: 1},
							{Content: "Center", Alignment: models.AlignCenter, RowSpan: 1, ColSpan: 1},
							{Content: "Right", Alignment: models.AlignRight, RowSpan: 1, ColSpan: 1},
						},
					},
					{
						Cells: []models.Cell{
							{Content: "1", Alignment: models.AlignNone, RowSpan: 1, ColSpan: 1},
							{Content: "2", Alignment: models.AlignNone, RowSpan: 1, ColSpan: 1},
							{Content: "3", Alignment: models.AlignNone, RowSpan: 1, ColSpan: 1},
						},
					},
				},
			},
			found: true,
		},
		{
			name: "nested formatting in cells",
			html: `<table><tr><td><b>Bold</b></td><td><i>Italic</i></td></tr></table>`,
			expected: models.Table{
				HasHeader: false,
				Rows: []models.Row{
					{
						Cells: []models.Cell{
							{Content: "**Bold**", Alignment: models.AlignNone, RowSpan: 1, ColSpan: 1},
							{Content: "_Italic_", Alignment: models.AlignNone, RowSpan: 1, ColSpan: 1},
						},
					},
				},
			},
			found: true,
		},
		{
			name: "confluence macro stripping",
			html: `<table><tr><td><span class="confluence-jim-macro confluence-status-lozenge" data-status-colour="GREEN">COMPLETE</span></td></tr></table>`,
			expected: models.Table{
				HasHeader: false,
				Rows: []models.Row{
					{
						Cells: []models.Cell{
							{Content: "COMPLETE", Alignment: models.AlignNone, RowSpan: 1, ColSpan: 1},
						},
					},
				},
			},
			found: true,
		},
		{
			name: "no table found",
			html: `<div>No table here</div>`,
			found: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, found := ExtractTableFromHTML(tt.html)
			assert.Equal(t, tt.found, found)
			if found {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
