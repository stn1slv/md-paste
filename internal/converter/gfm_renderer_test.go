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
