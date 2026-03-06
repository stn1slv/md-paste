<!--
Sync Impact Report:
- Version change: 0.1.0 → 0.1.1 (Clarified stdin/JSON scope)
- List of modified principles:
    - I. Unix Philosophy & Composability: Made stdin support optional/context-dependent.
    - III. Robust CLI Interface: Made JSON output optional/context-dependent.
- Added sections: None
- Removed sections: None
- Templates requiring updates: ✅ None (no structural change)
- Follow-up TODOs: None
-->

# md-paste Constitution

## Core Principles

### I. Unix Philosophy & Composability
The tool MUST do one thing well: handle Markdown content between the clipboard and the terminal. It SHOULD support stdin and stdout to enable pipe-based composition with other CLI tools when appropriate for the tool's domain (e.g., if stdin is a primary data source).

### II. Idiomatic Go Architecture
Follow the standard Go project layout: `cmd/` for entry points, `internal/` for private logic. NO logic in main packages. Business logic MUST be separated from CLI presentation.

### III. Robust CLI Interface
Provide a clear, self-documenting CLI using `cobra`. Support standard flags and environment variables. Outputs MUST be human-readable by default and SHOULD be machine-parsable (JSON) when specifically requested via flags for automation scenarios.

### IV. Quality Through TDD
TDD is mandatory: unit tests for all internal packages. Table-driven tests are the default for complex logic. Maintain 80%+ coverage for core business logic. Tests MUST fail before implementation is added.

### V. Idiomatic Error Handling
Errors MUST be treated as values and wrapped with context using `fmt.Errorf("...: %w", err)`. No panics are allowed except for critical startup failures. Use structured logging (`log/slog`) for diagnostics.

## Tooling & Automation
`Makefile` is the primary entry point for all development tasks (setup, test, lint, format, build). `golangci-lint` and `gofumpt` MUST pass before any PR is submitted. Use `uv` for any Python-based helper scripts.

## Governance & Stability
All changes MUST be peer-reviewed via Pull Requests. Semantic versioning (SemVer) MUST be followed. Breaking changes require a MAJOR version bump and a migration guide. This constitution takes precedence over all other local practices.

## Governance
This constitution supersedes all other informal practices. Amendments require a consensus review and an update to the `.specify/memory/constitution.md` file. All PRs and reviews MUST verify compliance with these principles. Use `.specify/memory/guidance.md` for runtime development hints and architecture decisions.

**Version**: 0.1.1 | **Ratified**: 2026-03-06 | **Last Amended**: 2026-03-06
