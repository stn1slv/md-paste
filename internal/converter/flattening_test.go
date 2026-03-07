package converter

import (
	"testing"

	"github.com/stn1slv/md-paste/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestFlattenTable(t *testing.T) {
	tests := []struct {
		name     string
		input    models.Table
		expected models.Table
	}{
		{
			name: "no merged cells",
			input: models.Table{
				Rows: []models.Row{
					{Cells: []models.Cell{{Content: "A"}, {Content: "B"}}},
					{Cells: []models.Cell{{Content: "C"}, {Content: "D"}}},
				},
			},
			expected: models.Table{
				Rows: []models.Row{
					{Cells: []models.Cell{{Content: "A", RowSpan: 1, ColSpan: 1}, {Content: "B", RowSpan: 1, ColSpan: 1}}},
					{Cells: []models.Cell{{Content: "C", RowSpan: 1, ColSpan: 1}, {Content: "D", RowSpan: 1, ColSpan: 1}}},
				},
			},
		},
		{
			name: "colspan",
			input: models.Table{
				Rows: []models.Row{
					{Cells: []models.Cell{{Content: "Merged", ColSpan: 2}}},
					{Cells: []models.Cell{{Content: "A"}, {Content: "B"}}},
				},
			},
			expected: models.Table{
				Rows: []models.Row{
					{Cells: []models.Cell{{Content: "Merged", RowSpan: 1, ColSpan: 1}, {Content: "Merged", RowSpan: 1, ColSpan: 1}}},
					{Cells: []models.Cell{{Content: "A", RowSpan: 1, ColSpan: 1}, {Content: "B", RowSpan: 1, ColSpan: 1}}},
				},
			},
		},
		{
			name: "rowspan",
			input: models.Table{
				Rows: []models.Row{
					{Cells: []models.Cell{{Content: "Merged", RowSpan: 2}, {Content: "A"}}},
					{Cells: []models.Cell{{Content: "B"}}},
				},
			},
			expected: models.Table{
				Rows: []models.Row{
					{Cells: []models.Cell{{Content: "Merged", RowSpan: 1, ColSpan: 1}, {Content: "A", RowSpan: 1, ColSpan: 1}}},
					{Cells: []models.Cell{{Content: "Merged", RowSpan: 1, ColSpan: 1}, {Content: "B", RowSpan: 1, ColSpan: 1}}},
				},
			},
		},
		{
			name: "complex rowspan and colspan",
			input: models.Table{
				Rows: []models.Row{
					{Cells: []models.Cell{{Content: "MergeBoth", RowSpan: 2, ColSpan: 2}, {Content: "A"}}},
					{Cells: []models.Cell{{Content: "B"}}},
				},
			},
			expected: models.Table{
				Rows: []models.Row{
					{Cells: []models.Cell{{Content: "MergeBoth", RowSpan: 1, ColSpan: 1}, {Content: "MergeBoth", RowSpan: 1, ColSpan: 1}, {Content: "A", RowSpan: 1, ColSpan: 1}}},
					{Cells: []models.Cell{{Content: "MergeBoth", RowSpan: 1, ColSpan: 1}, {Content: "MergeBoth", RowSpan: 1, ColSpan: 1}, {Content: "B", RowSpan: 1, ColSpan: 1}}},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FlattenTable(tt.input)
			assert.Equal(t, tt.expected.Rows, result.Rows)
		})
	}
}
