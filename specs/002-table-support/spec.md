# Feature Specification: Table Support for Word, PDF, and Confluence

**Feature Branch**: `002-table-support`  
**Created**: 2026-03-07  
**Status**: Draft  
**Input**: User description: "I want to add support of tables from Word, PDF and (also important) from Confluence pages"

## Clarifications

### Session 2026-03-07
- **Q**: Should Microsoft Excel be explicitly supported in the scope? → **A**: Explicitly support Excel alongside Word and Confluence using similar HTML parsing logic.
- **Q**: How should the Markdown table separator row handle cell alignment? → **A**: Standard GFM separator with alignment: Map source cell alignment to `:---`, `---:`, or `:---:`.
- **Q**: How should the system handle tables with multiple header rows (multiple rows of `<th>`)? → **A**: Single row only: Use the first `<tr>` containing `<th>` as the GFM header; treat others as regular rows.
- **Q**: How should the system handle tables with merged cells (rowspan/colspan)? → **A**: Flatten to GFM: Repeats the content of the merged cell across all its covered slots.
- **Q**: How deep should the PDF table reconstruction logic go? → **A**: Layout-Aware: Uses whitespace heuristics to guess columns and rows.
- **Q**: How should Confluence-specific macros (Status, User mentions, etc.) be represented? → **A**: Strip to Text: Converts macros to their plain text representation (e.g., `[In Progress]`).
- **Q**: If both structured HTML and plain text are present in the clipboard, which should be prioritized? → **A**: Strict Priority: Always prioritize `public.html` (structured) over plain text (heuristic).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Paste Table from Word, Excel, or Confluence (Priority: P1)

As a user, I want to copy a table from a Microsoft Word/Excel document or a Confluence page and paste it into my Markdown editor so that it is automatically converted into a GitHub Flavored Markdown (GFM) table.

**Why this priority**: Word, Excel, and Confluence are the primary sources of structured data in corporate environments. This delivers the core value for the majority of users.

**Independent Test**: Can be fully tested by copying a 3x3 table from any supported office application and verifying it appears as a valid GFM table in stdout/clipboard.

**Acceptance Scenarios**:

1. **Given** a 3x3 table in Word or Excel with a header row, **When** copied and pasted via `md-paste`, **Then** the output is a valid GFM table with the first row as the header and correct alignment markers (`:---`, etc.) in the separator.
2. **Given** a table in Confluence with cell alignments (left/right), **When** copied and pasted, **Then** the resulting GFM table preserves those alignments in the separator row.
3. **Given** a table with multiple header rows, **When** pasted, **Then** only the first row is used for the GFM header/separator boundary.

---

### User Story 2 - Paste Table from PDF (Priority: P2)

As a user, I want to copy a table from a PDF document (e.g., an invoice or report) and have it converted to a Markdown table, even if the clipboard data is primarily plain text.

**Why this priority**: PDF data is often "trapped" and manual reconstruction is tedious. This provides significant productivity gains.

**Independent Test**: Copy a structured table from a PDF viewer and verify the column structure is preserved in the Markdown output.

**Acceptance Scenarios**:

1. **Given** a multi-column table in a PDF, **When** copied and pasted, **Then** the system uses layout-aware heuristics to identify columns and produces a valid GFM table.
2. **Given** a PDF table with wrapped text in cells, **When** pasted, **Then** the system correctly identifies the row boundaries and merges wrapped lines into single cells.

---

### User Story 3 - Complex Table Formatting (Priority: P3)

As a user, I want complex table features like merged cells or nested formatting to be handled gracefully so that I don't lose information during conversion.

**Why this priority**: Ensures robustness for professional documents, though harder to achieve in pure Markdown.

**Independent Test**: Copy a table with merged cells and verify it is handled by flattening content across cells.

**Acceptance Scenarios**:

1. **Given** a table with a cell spanning two columns (A and B), **When** pasted, **Then** the output GFM table contains the content in both column A and column B for that row (flattening strategy).

---

### Edge Cases

- **Empty Cells**: How does the system handle tables where some cells are completely empty? (Expected: Empty cell `| |` in Markdown).
- **Nested Tables**: What happens if a table cell contains another table? (Expected: Flattened or ignored nested table).
- **Extremely Wide Tables**: How does the system handle tables that exceed standard Markdown readability? (Expected: Proceed with conversion, user handles wrapping).
- **Hidden Columns**: Should columns hidden in the source (e.g., Excel/Word) be included? (Expected: Skip hidden data if not present in clipboard).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST detect `public.html` (NSPasteboardTypeHTML) on macOS to extract structured table data from Word, Excel, and Confluence.
- **FR-002**: System MUST convert HTML `<table>`, `<tr>`, `<th>`, and `<td>` tags into GFM pipe-table syntax.
- **FR-003**: System MUST handle merged cells (rowspan/colspan) by **flattening** the content into each covered cell to preserve table structure in GFM.
- **FR-004**: System MUST reconstruct tables from plain text clipboard data (common in PDFs) using **layout-aware** heuristics (whitespace analysis) to identify column boundaries.
- **FR-005**: System MUST handle Confluence-specific elements (macros like Status, User Mentions) by **stripping them to their plain text representation** (e.g., `[IN PROGRESS]` becomes `IN PROGRESS`).
- **FR-006**: System MUST preserve bold/italic formatting within table cells.
- **FR-007**: System MUST escape pipe characters (`|`) within cell content to avoid breaking the GFM structure.
- **FR-008**: System MUST support Excel-specific HTML exports (e.g., detecting table structure in `mso-` prefixed HTML).
- **FR-009**: System MUST map CSS `text-align` properties (left, right, center) from source HTML to GFM alignment markers in the separator row.
- **FR-010**: System MUST only treat the first encountered row of table headers (`<th>`) as the GFM header row; subsequent header rows MUST be treated as standard body rows.
- **FR-011**: System MUST prioritize `public.html` (structured data) over `public.utf8-plain-text` (unstructured data) if both are present and contain table markers.

### Key Entities *(include if feature involves data)*

- **Table Data**: The raw representation of the copied table, containing rows, columns, and cell metadata (headers, alignment).
- **GFM Table**: The target representation using pipe and hyphen characters compatible with GitHub and most Markdown viewers.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 95% of standard Word/Excel/Confluence tables (no merged cells) are converted to valid GFM without data loss.
- **SC-002**: Reconstructed PDF tables preserve column integrity for at least 80% of common "text-selectable" PDF documents.
- **SC-003**: Conversion of a 100-cell table completes in under 100ms on a standard modern laptop.
- **SC-004**: Zero "broken" Markdown tables (missing pipes, mismatched columns) are generated from valid HTML table input.
