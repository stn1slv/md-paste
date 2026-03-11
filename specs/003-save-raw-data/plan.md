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
- [x] **II. Idiomatic Go**: Logic will be kept in `internal/cli/root.go` or a dedicated package.
- [x] **III. Robust CLI**: Uses `cobra` for the new flag, follows established silence-on-success.
- [x] **IV. Quality Through TDD**: Unit tests will verify file creation and content.
- [x] **V. Idiomatic Error Handling**: File system errors will be wrapped with context.

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

### Step 2: Implement File Saving Logic
- Create a helper function `saveRawContent(path string, content models.ClipboardContent) error`.
- Implement `os.Stat` check for directory.
- Implement priority logic: `RawHTML` > `PlainText`.
- Call `os.WriteFile` with `0644` permissions.

### Step 3: Integrate into `runPaste`
- Call `saveRawContent` after `clipboardRead()` and before `converter.Convert()`.
- Ensure it only runs if `saveRawFlag != ""`.
- Ensure errors are handled according to the spec (return error and exit).

## Phase 3: Verification Plan

### Unit Tests
- Test `saveRawContent` with various paths (existing file, new file, directory, unwritable path).
- Verify file content matches input `ClipboardContent`.
- Mock filesystem using a temporary directory for tests.

### Integration Tests
- Verify `--save-raw` works correctly alongside `--stdout`.
- Verify silence-on-success behavior.
- Verify error output when providing a directory path.
