# Tasks: Table Support

This document defines the actionable tasks for implementing table support from Word, Excel, Confluence, and PDF.

## Implementation Strategy

We follow an incremental delivery approach, prioritizing high-fidelity HTML-based sources first (Word/Excel/Confluence) followed by layout-aware PDF reconstruction. Each phase results in an independently testable and useful increment.

1. **Foundational**: Define data models and rendering logic.
2. **Word/Excel/Confluence (P1)**: Implement HTML table parsing and alignment mapping.
3. **PDF (P2)**: Implement layout-aware plain text reconstruction.
4. **Complex Formatting (P3)**: Implement cell flattening for merged cells.

## Phase 1: Setup

Initial environment and dependency setup.

- [x] T001 Add `golang.org/x/net/html` dependency via `go get golang.org/x/net/html && go mod tidy`

## Phase 2: Foundational

Core data structures and GFM rendering logic that all user stories depend on.

- [x] T002 Define `Table`, `Row`, `Cell` structs and `Alignment` enum in `internal/models/models.go`
- [x] T003 Implement GFM table rendering logic in `internal/converter/gfm_renderer.go`
- [x] T004 [P] Create unit tests for GFM rendering in `internal/converter/gfm_renderer_test.go`

## Phase 3: User Story 1 - Word, Excel, or Confluence (Priority: P1)

**Goal**: Automatically convert structured HTML tables copied from Word, Excel, or Confluence.
**Independent Test**: Copy a simple 3x3 table from any supported office application and verify it outputs a valid GFM pipe table via `md-paste`.

- [x] T005 [US1] Implement HTML table extraction logic using `golang.org/x/net/html` in `internal/converter/html_table.go`
- [x] T006 [US1] Implement cell alignment extraction from `style` or `align` attributes in `internal/converter/html_table.go`
- [x] T007 [US1] Implement Confluence macro stripping (extracting inner text) in `internal/converter/html_table.go`
- [x] T008 [P] [US1] Create unit tests with HTML table samples in `internal/converter/html_table_test.go`
- [x] T009 [US1] Integrate HTML table parsing into `internal/converter/converter.go`'s `Convert` function

## Phase 4: User Story 2 - PDF Reconstruction (Priority: P2)

**Goal**: Reconstruct table structure from plain text copied from PDF using layout-aware heuristics.
**Independent Test**: Copy a text block with multiple spaces/tabs separating columns and verify it is converted to a valid GFM table.

- [x] T010 [US2] Implement layout-aware text parser using whitespace heuristics in `internal/converter/text_table.go`
- [x] T011 [P] [US2] Create unit tests for text table reconstruction in `internal/converter/text_table_test.go`
- [x] T012 [US2] Integrate text table parsing as a fallback in `internal/converter/converter.go`

## Phase 5: User Story 3 - Complex Table Formatting (Priority: P3)

**Goal**: Support merged cells (rowspan/colspan) by flattening content across the grid.
**Independent Test**: Copy a table with merged cells and verify the content repeats in all spanned GFM cells.

- [x] T013 [US3] Update `Cell` model to include `RowSpan` and `ColSpan` in `internal/models/models.go`
- [x] T014 [US3] Implement table grid flattening logic (injecting duplicate cells) in `internal/converter/flattening.go`
- [x] T015 [P] [US3] Create unit tests for table flattening in `internal/converter/flattening_test.go`
- [x] T016 [US3] Update `html_table.go` to populate `RowSpan` and `ColSpan` data during parsing

## Phase 6: Polish & Cross-Cutting Concerns

Ensuring robustness and consistency across the entire feature.

- [x] T017 Ensure graceful fallback to standard `html-to-markdown` conversion if no table is detected in `internal/converter/converter.go`
- [x] T018 Run `make lint` and `make format` to ensure code quality across all new files
- [x] T019 Update `internal/cli/root_test.go` to include a table conversion integration test case

## Dependencies

- All User Stories depend on Phase 2 (Foundational).
- US1 (P1) is the MVP and should be implemented first.
- US2 (P2) can be implemented independently after Phase 2.
- US3 (P3) depends on US1 (HTML parsing).

## Parallel Execution Examples

### User Story 1 (P1) Parallel Paths
- T005, T006, T007 can be developed together in `html_table.go`.
- T008 (Tests) can be developed in parallel with T005-T007 if samples are provided.

### User Story 2 (P2) Parallel Path
- T010 (Parser) and T011 (Tests) can be developed in parallel by different contributors.
