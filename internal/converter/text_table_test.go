package converter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stn1slv/md-paste/internal/models"
)

func TestExtractTableFromText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected models.Table
		found    bool
	}{
		{
			name: "simple space separated table",
			text: "Header1    Header2\nData1      Data2",
			expected: models.Table{
				Rows: []models.Row{
					{
						Cells: []models.Cell{
							{Content: "Header1", Alignment: models.AlignNone, RowSpan: 1, ColSpan: 1},
							{Content: "Header2", Alignment: models.AlignNone, RowSpan: 1, ColSpan: 1},
						},
					},
					{
						Cells: []models.Cell{
							{Content: "Data1", Alignment: models.AlignNone, RowSpan: 1, ColSpan: 1},
							{Content: "Data2", Alignment: models.AlignNone, RowSpan: 1, ColSpan: 1},
						},
					},
				},
			},
			found: true,
		},
		{
			name: "tab separated table",
			text: "H1\tH2\nD1\tD2",
			expected: models.Table{
				Rows: []models.Row{
					{
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
			name: "irregular spacing",
			text: "Col1  Col2    Col3\nVal1    Val2  Val3",
			expected: models.Table{
				Rows: []models.Row{
					{
						Cells: []models.Cell{
							{Content: "Col1", Alignment: models.AlignNone, RowSpan: 1, ColSpan: 1},
							{Content: "Col2", Alignment: models.AlignNone, RowSpan: 1, ColSpan: 1},
							{Content: "Col3", Alignment: models.AlignNone, RowSpan: 1, ColSpan: 1},
						},
					},
					{
						Cells: []models.Cell{
							{Content: "Val1", Alignment: models.AlignNone, RowSpan: 1, ColSpan: 1},
							{Content: "Val2", Alignment: models.AlignNone, RowSpan: 1, ColSpan: 1},
							{Content: "Val3", Alignment: models.AlignNone, RowSpan: 1, ColSpan: 1},
						},
					},
				},
			},
			found: true,
		},
		{
			name: "max columns normalization (header has fewer)",
			text: "H1  H2\nD1  D2  D3",
			expected: models.Table{
				Rows: []models.Row{
					{
						Cells: []models.Cell{
							{Content: "H1", Alignment: models.AlignNone, RowSpan: 1, ColSpan: 1},
							{Content: "H2", Alignment: models.AlignNone, RowSpan: 1, ColSpan: 1},
							{Content: "", Alignment: models.AlignNone, RowSpan: 1, ColSpan: 1},
						},
					},
					{
						Cells: []models.Cell{
							{Content: "D1", Alignment: models.AlignNone, RowSpan: 1, ColSpan: 1},
							{Content: "D2", Alignment: models.AlignNone, RowSpan: 1, ColSpan: 1},
							{Content: "D3", Alignment: models.AlignNone, RowSpan: 1, ColSpan: 1},
						},
					},
				},
			},
			found: true,
		},
		{
			name:  "not a table (single column)",
			text:  "Just some text\nOn multiple lines",
			found: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, found := ExtractTableFromText(tt.text)
			assert.Equal(t, tt.found, found)
			if found {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
