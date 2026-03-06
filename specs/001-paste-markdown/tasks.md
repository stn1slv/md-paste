---

description: "Task list for Paste Markdown CLI implementation"
---

# Tasks: Paste Markdown CLI (001-paste-markdown)

**Input**: Design documents from `/specs/001-paste-markdown/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: TDD is MANDATORY per project constitution (IV. Quality Through TDD). Write tests first and ensure they fail.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)

## Path Conventions

- **Single project**: `cmd/`, `internal/`, `tests/` at repository root

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [x] T001 [P] Create directory structure: `cmd/md-paste`, `internal/clipboard`, `internal/converter`, `internal/cli`, `internal/models`, `tests/unit`, `tests/integration`
- [x] T002 [P] Update `go.mod` with dependencies: `cobra`, `html-to-markdown`, `testify`
- [x] T003 [P] Configure `.golangci.yml` with revive, gosec, staticcheck, etc., per spec
- [x] T004 [P] Verify `Makefile` targets: `setup`, `test`, `lint`, `format`, `build`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

- [x] T005 Define `ClipboardContent` and `MarkdownDocument` structs in `internal/models/models.go`
- [x] T006 Setup base logging utility in `internal/logger/logger.go` (using `log/slog`)
- [x] T007 [P] Implement error wrapping utility in `internal/errors/errors.go` (fmt.Errorf with %w)

**Checkpoint**: Foundation ready - user story implementation can now begin

---

## Phase 3: User Story 1 - Convert Clipboard HTML to Markdown (Priority: P1) 🎯 MVP

**Goal**: Copy rich-text/HTML, run `md-paste`, and get Markdown on clipboard.

**Independent Test**: Copy HTML from browser, run `md-paste`, paste into editor, verify Markdown.

### Tests for User Story 1 (MANDATORY - TDD) ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T008 [P] [US1] Unit tests for clipboard reading (HTML/Plain fallback) in `internal/clipboard/clipboard_test.go`
- [x] T009 [P] [US1] Unit tests for HTML to Markdown conversion in `internal/converter/converter_test.go`
- [x] T010 [P] [US1] Unit tests for clipboard writing in `internal/clipboard/clipboard_test.go`
- [x] T011 [US1] Integration tests for Safari, Chrome, and MS Word clipboard formats in `tests/integration/clipboard_flow_test.go`

### Implementation for User Story 1

- [x] T012 [US1] Implement native macOS `NSPasteboard` reader (CGO) in `internal/clipboard/clipboard_darwin.go`
- [x] T013 [US1] Implement HTML to Markdown conversion using `html-to-markdown` in `internal/converter/converter.go`
- [x] T014 [US1] Implement native macOS `NSPasteboard` writer (CGO) in `internal/clipboard/clipboard_darwin.go`
- [x] T015 [US1] Create root Cobra command in `internal/cli/root.go` to orchestrate US1 flow
- [x] T016 [US1] Implement main entry point in `cmd/md-paste/main.go`
- [x] T017 [US1] Add validation for non-text/empty clipboard data in `internal/clipboard/clipboard_darwin.go`

**Checkpoint**: User Story 1 fully functional and testable independently. MVP achieved.

---

## Phase 4: User Story 2 - Output Markdown to Stdout (Priority: P2)

**Goal**: Support `--stdout` flag to print result instead of writing to clipboard.

**Independent Test**: Run `md-paste --stdout` and verify output is in terminal and clipboard is unchanged.

### Tests for User Story 2 (MANDATORY - TDD) ⚠️

- [x] T018 [P] [US2] Unit tests for CLI flag parsing and stdout redirection logic in `internal/cli/root_test.go`
- [x] T019 [US2] Integration test for `--stdout` output flow in `tests/integration/stdout_flow_test.go`

### Implementation for User Story 2

- [x] T020 [US2] Add `--stdout` / `-s` flag to Cobra command in `internal/cli/root.go`
- [x] T021 [US2] Implement conditional logic for Stdout vs Clipboard output in `internal/cli/root.go`
- [x] T022 [US2] Ensure "Silence-on-Success" for the clipboard write path (FR-005.1)

**Checkpoint**: User Story 2 functional. Tool supports Unix pipes and redirection.

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Final improvements and quality gates

- [x] T023 [P] Add `--help` documentation with examples in `internal/cli/root.go`
- [x] T024 [P] Final lint and format check: `make lint`, `make format`
- [x] T025 [P] Verify SC-002: Execution time < 200ms for 100KB payloads
- [x] T026 [P] Update `README.md` with usage instructions and installation steps

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - starts immediately.
- **Foundational (Phase 2)**: Depends on Phase 1 completion.
- **User Story 1 (Phase 3)**: Depends on Phase 2 completion. BLOCKS Phase 4.
- **User Story 2 (Phase 4)**: Depends on Phase 3 completion.
- **Polish (Final Phase)**: Depends on Phase 3 & 4 completion.

### User Story Completion Order

1. **US1 (MVP)**: Core Read -> Convert -> Write flow.
2. **US2**: Extended flag support for Stdout.

### Parallel Opportunities

- T001, T002, T003 (Setup phase)
- T007, T008, T009 (US1 Unit Tests)
- T017 (US2 Unit Tests)
- T022, T023, T025 (Polish phase)

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1 & 2.
2. Complete Phase 3 (US1).
3. **STOP and VALIDATE**: Copy HTML, run `md-paste`, verify editor paste.

### Incremental Delivery

1. Foundation ready.
2. US1 complete -> MVP delivered.
3. US2 complete -> Pipe support added.
4. Polish complete -> Final release.

### Parallel Team Strategy

- Developer A: Implement `internal/clipboard` (T007, T011, T009, T013)
- Developer B: Implement `internal/converter` (T008, T012)
- Developer C: Implement CLI structure and orchestration (T014, T015, T019, T020)

---

## Notes

- TDD is strictly enforced per Constitution. No implementation before failing tests.
- CGO is required for `NSPasteboard`.
- JSON output and stdin support are NOT required for this implementation.
- Silence-on-Success is the default behavior for clipboard operations.
