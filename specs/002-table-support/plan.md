# Implementation Plan: Table Support

**Branch**: `002-table-support` | **Date**: 2026-03-07 | **Spec**: [specs/002-table-support/spec.md](spec.md)
**Input**: Feature specification from `/specs/002-table-support/spec.md`

## Summary

Add support for automatically converting tables copied from Word, Excel, Confluence, and PDF documents into GitHub Flavored Markdown (GFM) tables. The technical approach involves utilizing `golang.org/x/net/html` to parse structured `public.html` clipboards, generating an intermediate table model to flatten merged cells and detect alignments, and using layout-aware heuristics to reconstruct tables from plain text clipboard data.

## Technical Context

**Language/Version**: Go 1.26+  
**Primary Dependencies**: `golang.org/x/net/html`  
**Storage**: N/A
**Testing**: `testify/assert`, `testify/require`  
**Target Platform**: macOS (specifically utilizing `public.html` on `NSPasteboard`), fallback to standard behavior on other OSes.  
**Project Type**: CLI  
**Performance Goals**: < 100ms processing time for a 100-cell table.  
**Constraints**: Must strictly output valid GFM, no panics on malformed HTML, handle nested formatting.  
**Scale/Scope**: Local execution.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **I. Unix Philosophy & Composability**: PASS - The tool transparently enhances clipboard data flow without introducing interactive prompts.
- **II. Idiomatic Go Architecture**: PASS - Logic will be built into an `internal/converter` subpackage or similar, separate from the `cmd` package.
- **III. Robust CLI Interface**: PASS - No new flags are strictly necessary; integrates natively into standard pipeline.
- **IV. Quality Through TDD**: PASS - Unit tests will be required for the HTML parsing, text heuristics, and GFM rendering stages.
- **V. Idiomatic Error Handling**: PASS - Fallback behavior ensures the program outputs plain text/markdown without erroring out if table parsing fails.

## Project Structure

### Documentation (this feature)

```text
specs/002-table-support/
├── plan.md              
├── research.md          
├── data-model.md        
├── quickstart.md        
├── contracts/           
└── tasks.md             
```

### Source Code (repository root)

```text
src/
├── cmd/
│   └── md-paste/
├── internal/
│   ├── converter/
│   │   ├── html_table.go
│   │   ├── html_table_test.go
│   │   ├── text_table.go
│   │   └── text_table_test.go
│   ├── clipboard/
│   └── logger/
```

**Structure Decision**: Retain the single project structure following standard Go project layout, placing parsing logic within `internal/converter`.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| N/A | N/A | N/A |
