package converter

import (
	"testing"

	"github.com/stn1slv/md-paste/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestRenderTable(t *testing.T) {
	tests := []struct {
		name     string
		table    models.Table
		expected string
	}{
		{
			name: "simple 2x2 table",
			table: models.Table{
				Rows: []models.Row{
					{
						Cells: []models.Cell{
							{Content: "Header 1"},
							{Content: "Header 2"},
						},
					},
					{
						Cells: []models.Cell{
							{Content: "Data 1"},
							{Content: "Data 2"},
						},
					},
				},
			},
			expected: "| Header 1 | Header 2 |\n| --- | --- |\n| Data 1 | Data 2 |",
		},
		{
			name: "table with alignment",
			table: models.Table{
				Rows: []models.Row{
					{
						Cells: []models.Cell{
							{Content: "Left", Alignment: models.AlignLeft},
							{Content: "Center", Alignment: models.AlignCenter},
							{Content: "Right", Alignment: models.AlignRight},
						},
					},
					{
						Cells: []models.Cell{
							{Content: "1"},
							{Content: "2"},
							{Content: "3"},
						},
					},
				},
			},
			expected: "| Left | Center | Right |\n| :--- | :---: | ---: |\n| 1 | 2 | 3 |",
		},
		{
			name: "escaping pipes",
			table: models.Table{
				Rows: []models.Row{
					{
						Cells: []models.Cell{
							{Content: "Header | with pipe"},
						},
					},
					{
						Cells: []models.Cell{
							{Content: "Data | with pipe"},
						},
					},
				},
			},
			expected: "| Header \\| with pipe |\n| --- |\n| Data \\| with pipe |",
		},
		{
			name: "sanitizing newlines",
			table: models.Table{
				Rows: []models.Row{
					{
						Cells: []models.Cell{
							{Content: "Line 1\nLine 2"},
						},
					},
				},
			},
			expected: "| Line 1 Line 2 |\n| --- |",
		},
		{
			// GFM forces the header to be the first row. A table whose marked
			// header is not the first row must keep its row order; only the
			// separator alignment is borrowed from the marked row.
			name: "header row after the first row does not reorder rows",
			table: models.Table{
				HasHeader: true,
				Rows: []models.Row{
					{
						Cells: []models.Cell{
							{Content: "Caption A"},
							{Content: "Caption B"},
						},
					},
					{
						IsHeader: true,
						Cells: []models.Cell{
							{Content: "Head A", Alignment: models.AlignRight},
							{Content: "Head B", Alignment: models.AlignCenter},
						},
					},
					{
						Cells: []models.Cell{
							{Content: "Data A"},
							{Content: "Data B"},
						},
					},
				},
			},
			expected: "| Caption A | Caption B |\n| ---: | :---: |\n| Head A | Head B |\n| Data A | Data B |",
		},
		{
			name: "empty table",
			table: models.Table{
				Rows: []models.Row{},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderTable(tt.table)
			assert.Equal(t, tt.expected, result)
		})
	}
}
