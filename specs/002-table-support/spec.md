# Feature Specification: Table Support for Word, PDF, and Confluence

**Feature Branch**: `002-table-support`  
**Created**: 2026-03-07  
**Status**: Draft  
**Input**: User description: "I want to add support of tables from Word, PDF and (also important) from Confluence pages"

## Clarifications

### Session 2026-03-07
- Q: Should Microsoft Excel be explicitly supported in the scope? → A: Explicitly support Excel alongside Word and Confluence using similar HTML parsing logic.
- Q: How should the Markdown table separator row handle cell alignment? → A: Standard GFM separator with alignment: Map source cell alignment to `:---`, `---:`, or `:---:`.
- Q: How should the system handle tables with multiple header rows (multiple rows of `<th>`)? → A: Single row only: Use the first `<tr>` containing `<th>` as the GFM header; treat others as regular rows.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Paste Table from Word, Excel, or Confluence (Priority: P1)

...

**Independent Test**: Can be fully tested by copying a 3x3 table from any supported office application and verifying it appears as a valid GFM table in stdout/clipboard.

**Acceptance Scenarios**:

1. **Given** a 3x3 table in Word or Excel with a header row, **When** copied and pasted via `md-paste`, **Then** the output is a valid GFM table with the first row as the header and correct alignment markers (`:---`, etc.) in the separator.
2. **Given** a table with multiple header rows, **When** pasted, **Then** only the first row is used for the GFM header/separator boundary.

---

### User Story 2 - Paste Table from PDF (Priority: P2)
...

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST detect `public.html` (NSPasteboardTypeHTML) on macOS to extract structured table data from Word, Excel, and Confluence.
- **FR-002**: System MUST convert HTML `<table>`, `<tr>`, `<th>`, and `<td>` tags into GFM pipe-table syntax.
...
- **FR-008**: System MUST support Excel-specific HTML exports (e.g., detecting table structure in `mso-` prefixed HTML).
- **FR-009**: System MUST map CSS `text-align` properties (left, right, center) from source HTML to GFM alignment markers in the separator row.
- **FR-010**: System MUST only treat the first encountered row of table headers (`<th>`) as the GFM header row; subsequent header rows MUST be treated as standard body rows.

### Key Entities *(include if feature involves data)*
...