# Implementation Plan: Save raw clipboard data to file

**Feature Branch**: `003-save-raw-data`  
**Created**: 2026-03-11  
**Status**: Draft  
**Reference Spec**: [specs/003-save-raw-data/spec.md]

## Technical Context

| Area | Choice | Rationale |
|------|--------|-----------|
| **Language** | Go 1.26+ | Project standard. |
| **CLI Framework** | Cobra | Project standard. |
| **File I/O** | `os.WriteFile` | Atomic-style writing, handles truncation. |
| **Validation** | `os.Stat` | Explicit directory check for better error reporting. |

## Constitution Check

- [x] **I. Unix Philosophy**: Feature enables composability by saving raw source for audit/debug.
- [x] **II. Idiomatic Go**: Business logic is separated into `internal/clipboard/exporter.go`.
- [x] **III. Robust CLI**: Uses `cobra` for the new flag, follows established silence-on-success.
- [x] **IV. Quality Through TDD**: Unit tests in `internal/clipboard/exporter_test.go` verify file creation.
- [x] **V. Idiomatic Error Handling**: File system errors are wrapped with context.

## Phase 0: Research & Outline

- Decisions documented in [specs/003-save-raw-data/research.md].
- All "NEEDS CLARIFICATION" resolved in spec.

## Phase 1: Design & Contracts

- Data model: [specs/003-save-raw-data/data-model.md].
- CLI Contract: [specs/003-save-raw-data/contracts/cli.md].
- Quickstart: [specs/003-save-raw-data/quickstart.md].
- Agent context updated via `update-agent-context.sh`.

## Phase 2: Implementation Strategy

### Step 1: Update CLI Flags
- Add `saveRawFlag` string variable to `internal/cli/root.go`.
- Bind `--save-raw` / `-r` flag in `init()`.

### Step 2: Implement Export Logic
- Create `internal/clipboard/exporter.go`.
- Implement `SaveRaw(path string, content models.ClipboardContent) error`.
- Include `os.Stat` check for directory and permission validation.
- Implement priority logic: `RawHTML` > `PlainText`.
- Call `os.WriteFile` with `0644` permissions.

### Step 3: Integrate into `runPaste`
- In `internal/cli/root.go`, call `exporter.SaveRaw` after `clipboardRead()`.
- Ensure it only runs if `saveRawFlag != ""` AND the clipboard is not empty (`FR-004`).
- Ensure errors are handled according to the spec (return wrapped error and exit).

## Phase 3: Verification Plan

### Unit Tests
- Create `internal/clipboard/exporter_test.go`.
- Test `SaveRaw` with various paths (existing file, new file, directory, unwritable path).
- Verify file content matches input `ClipboardContent` priority.

### Integration Tests
- Update `tests/integration/stdout_flow_test.go` to verify `--save-raw` alongside `--stdout`.
- Verify silence-on-success behavior.
- Verify error output when providing an invalid path.
