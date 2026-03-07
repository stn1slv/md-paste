# Data Model: Table Support

This document outlines the internal domain model used to represent and process tables before rendering them as GitHub Flavored Markdown (GFM).

## Entities

### `Table`

Represents a complete table structure extracted from either HTML or plain text.

**Fields**:
- `Rows` (`[]Row`): Ordered collection of rows. The first row may be treated as the header.
- `HasHeader` (`bool`): Indicates if a distinct header row was detected (e.g., from `<th>` tags).

### `Row`

Represents a single row within a table.

**Fields**:
- `Cells` (`[]Cell`): Ordered collection of cells in this row.
- `IsHeader` (`bool`): True if this row was determined to be a header row.

### `Cell`

Represents an individual cell's data and formatting.

**Fields**:
- `Content` (`string`): The inner text of the cell. If the cell contains nested formatting (like bold/italic), this content is the already-converted Markdown representation of that inner HTML.
- `Alignment` (`Alignment`): The alignment for this column, derived from CSS `text-align` or HTML attributes (`align`). Defaults to `AlignNone`.
- `RowSpan` (`int`): Number of rows this cell spans (default 1).
- `ColSpan` (`int`): Number of columns this cell spans (default 1).

### `Alignment` (Enum)

- `AlignNone`: No specific alignment (Markdown default `---`).
- `AlignLeft`: Left-aligned (`:---`).
- `AlignCenter`: Center-aligned (`:---:`).
- `AlignRight`: Right-aligned (`---:`).

## State Transitions & Processing Flow

1. **Extraction**: 
   - From `public.html`: Parse `<table>` DOM into the `Table` struct. Read `rowspan` and `colspan`.
   - From plain text: Split lines on `\n`. Split columns on `\s{2,}` or `\t`. Build `Table` struct with 1x1 cells.
2. **Flattening**:
   - Iterate over `Table`. For any `Cell` with `RowSpan > 1` or `ColSpan > 1`, inject duplicate cells with the same `Content` into the subsequent rows/columns covered by the span, ensuring the table becomes a perfect grid.
3. **Rendering**:
   - Iterate over the grid-perfect `Table`.
   - Escape pipe `|` characters in `Content` to `\|`.
   - Output Row 0.
   - Output Separator Row based on `Cell.Alignment` of Row 0 (or default).
   - Output Row 1..N.
