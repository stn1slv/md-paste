package converter

import (
	"testing"

	"github.com/stn1slv/md-paste/internal/models"
	"github.com/stretchr/testify/assert"
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
			name: "empty cells in the middle",
			text: "Col1    Col2    Col3\nVal1            Val3",
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
							{Content: "Val3", Alignment: models.AlignNone, RowSpan: 1, ColSpan: 1},
							{Content: "", Alignment: models.AlignNone, RowSpan: 1, ColSpan: 1},
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
		{
			// An indented line must not gain a leading empty column from the
			// indentation matching the column separator.
			name: "indented rows keep their column count",
			text: "Col1  Col2\n\tVal1  Val2",
			expected: models.Table{
				Rows: []models.Row{
					{
						Cells: []models.Cell{
							{Content: "Col1", Alignment: models.AlignNone, RowSpan: 1, ColSpan: 1},
							{Content: "Col2", Alignment: models.AlignNone, RowSpan: 1, ColSpan: 1},
						},
					},
					{
						Cells: []models.Cell{
							{Content: "Val1", Alignment: models.AlignNone, RowSpan: 1, ColSpan: 1},
							{Content: "Val2", Alignment: models.AlignNone, RowSpan: 1, ColSpan: 1},
						},
					},
				},
			},
			found: true,
		},
		{
			name:  "indented prose is not a table",
			text:  "Just some text\n  On multiple lines",
			found: false,
		},
		{
			// A single line with a double space inside a block of prose is not
			// enough evidence of a table.
			name:  "one double-spaced line in prose is not a table",
			text:  "This is a sentence.  Next one here.\nAnother line follows.",
			found: false,
		},
		{
			// Rows that disagree wildly on their column count are prose split by
			// incidental whitespace, not a table.
			name:  "ragged column counts are not a table",
			text:  "one  two\nthree  four  five\nsix  seven  eight  nine\nten  eleven  twelve  thirteen  fourteen",
			found: false,
		},
		{
			name:  "numbered list with tabs (should NOT be a table)",
			text:  "1.\tFirst item\n2.\tSecond item",
			found: false,
		},
		{
			name:  "bullet list with tabs (should NOT be a table)",
			text:  "•\tFirst item\n•\tSecond item",
			found: false,
		},
		{
			name: "numbered 3-column table (SHOULD still be a table)",
			text: "No.\tItem\tPrice\n1.\tApple\t$1\n2.\tBanana\t$2",
			expected: models.Table{
				Rows: []models.Row{
					{
						Cells: []models.Cell{
							{Content: "No.", Alignment: models.AlignNone, RowSpan: 1, ColSpan: 1},
							{Content: "Item", Alignment: models.AlignNone, RowSpan: 1, ColSpan: 1},
							{Content: "Price", Alignment: models.AlignNone, RowSpan: 1, ColSpan: 1},
						},
					},
					{
						Cells: []models.Cell{
							{Content: "1.", Alignment: models.AlignNone, RowSpan: 1, ColSpan: 1},
							{Content: "Apple", Alignment: models.AlignNone, RowSpan: 1, ColSpan: 1},
							{Content: "$1", Alignment: models.AlignNone, RowSpan: 1, ColSpan: 1},
						},
					},
					{
						Cells: []models.Cell{
							{Content: "2.", Alignment: models.AlignNone, RowSpan: 1, ColSpan: 1},
							{Content: "Banana", Alignment: models.AlignNone, RowSpan: 1, ColSpan: 1},
							{Content: "$2", Alignment: models.AlignNone, RowSpan: 1, ColSpan: 1},
						},
					},
				},
			},
			found: true,
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
