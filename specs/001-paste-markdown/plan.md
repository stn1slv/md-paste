# Implementation Plan: Paste Markdown CLI (001-paste-markdown)

**Branch**: `001-paste-markdown` | **Date**: 2026-03-06 | **Spec**: [specs/001-paste-markdown/spec.md](specs/001-paste-markdown/spec.md)
**Input**: Feature specification from `/specs/001-paste-markdown/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

This feature implements a CLI tool, `md-paste`, that reads rich text (HTML) from the macOS clipboard using `NSPasteboard` via CGO, converts it to Markdown using the `html-to-markdown` library, and writes it back to the clipboard. It supports an optional `--stdout` flag for piping output to other tools.

## Technical Context

**Language/Version**: Go 1.26+  
**Primary Dependencies**: `cobra`, `html-to-markdown`, `testify/assert`, `testify/require`  
**Storage**: N/A  
**Testing**: `testing` package with Table-Driven Tests  
**Target Platform**: macOS  
**Project Type**: CLI  
**Performance Goals**: <200ms p95 for 100KB payloads  
**Constraints**: macOS native integration (`NSPasteboard` via CGO)  
**Scale/Scope**: Local clipboard operations only.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Unix Philosophy**: `md-paste` does one thing well (clipboard-to-clipboard or clipboard-to-stdout). Supports pipe via `--stdout`. (Pass)
- **Idiomatic Go Architecture**: Uses `cmd/` for entry points and `internal/` for logic. (Pass)
- **Robust CLI Interface**: Powered by `cobra`, supports flags and help. (Pass)
- **Quality Through TDD**: Mandatory testing for internal packages. (Pass)
- **Idiomatic Error Handling**: Using `fmt.Errorf` with `%w` and `slog`. (Pass)

## Project Structure

### Documentation (this feature)

```text
specs/001-paste-markdown/
├── plan.md              # This file
├── research.md          # Research findings
├── data-model.md        # Entities: ClipboardContent, MarkdownDocument
├── quickstart.md        # Guide for setup and usage
├── contracts/           # CLI contract (flags, outputs, exit codes)
└── tasks.md             # To be created by /speckit.tasks
```

### Source Code (repository root)

```text
# Option 1: Single project (DEFAULT)
cmd/
└── md-paste/
    └── main.go

internal/
├── clipboard/           # macOS native integration (CGO)
├── converter/           # HTML to Markdown logic (using html-to-markdown)
└── cli/                 # Cobra command structure

tests/
├── unit/                # Unit tests for internal packages
└── integration/         # Integration tests for end-to-end flow
```

**Structure Decision**: Single project layout as defined in the Go Application Blueprint.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| CGO Integration | Mandated by spec for native macOS NSPasteboard | `atotto/clipboard` is purely text-based and lacks HTML support. |
