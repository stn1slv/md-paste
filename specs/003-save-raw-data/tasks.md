---
description: "Task list for the save-raw-data feature implementation"
---

# Tasks: Save raw clipboard data to file

**Input**: Design documents from `/specs/003-save-raw-data/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Tests are included as per the verification plan in plan.md.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure verification

- [ ] T001 Verify `ClipboardContent` model in `internal/models/models.go` supports `RawHTML` and `PlainText`
- [ ] T002 [P] Review `internal/errors/errors.go` for appropriate wrapping of filesystem errors

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [ ] T003 Define `saveRawFlag` string variable for flag binding in `internal/cli/root.go`
- [ ] T004 Bind `--save-raw` / `-r` flag to `saveRawFlag` in `init()` in `internal/cli/root.go`

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Debugging Rich Text (Priority: P1) 🎯 MVP

**Goal**: Save raw HTML (prioritized) or PlainText from clipboard to a user-specified file to assist in debugging.

**Independent Test**: Copy HTML from a browser, run `md-paste --save-raw out.html`, and verify `out.html` contains the exact raw HTML data from the clipboard.

### Tests for User Story 1

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T005 [P] [US1] Create unit tests for `SaveRaw` logic in `internal/clipboard/exporter_test.go`
- [ ] T006 [P] [US1] Add test cases for directory paths and permission denial errors in `internal/clipboard/exporter_test.go`

### Implementation for User Story 1

- [ ] T007 [US1] Implement `SaveRaw(path string, content models.ClipboardContent) error` helper in `internal/clipboard/exporter.go`
- [ ] T008 [US1] Add directory check using `os.Stat` in `internal/clipboard/exporter.go`
- [ ] T009 [US1] Implement content priority (HTML > PlainText) and `os.WriteFile` in `internal/clipboard/exporter.go`
- [ ] T010 [US1] Ensure `runPaste` in `internal/cli/root.go` skips `SaveRaw` if clipboard is empty (FR-004)
- [ ] T011 [US1] Integrate `exporter.SaveRaw` call into `runPaste` flow before conversion in `internal/cli/root.go`

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently.

---

## Phase 4: User Story 2 - Integrated Export Pipeline (Priority: P2)

**Goal**: Support simultaneous raw data saving and converted Markdown output to standard output.

**Independent Test**: Run `md-paste --save-raw raw.html --stdout > out.md` and verify both `raw.html` (original) and `out.md` (converted) are created correctly.

### Tests for User Story 2

- [ ] T012 [P] [US2] Add integration test for combined `--save-raw` and `--stdout` usage in `tests/integration/stdout_flow_test.go`

### Implementation for User Story 2

- [ ] T013 [US2] Verify `runPaste` flow correctly handles both `saveRawFlag` and `stdoutFlag` in `internal/cli/root.go`
- [ ] T014 [US2] Ensure errors in `exporter.SaveRaw` correctly terminate the pipeline even when `--stdout` is used in `internal/cli/root.go`

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently.

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T015 Verify silence-on-success behavior when using `--save-raw` without `--stdout` in `internal/cli/root.go`
- [ ] T016 [P] Run `make lint` and `make test` to validate implementation across the project
- [ ] T017 [P] Run `specs/003-save-raw-data/quickstart.md` validation scenarios

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately.
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories.
- **User Stories (Phase 3+)**: All depend on Foundational phase completion.
- **Polish (Final Phase)**: Depends on all desired user stories being complete.

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories.
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - Integrates with US1 logic but can be tested independently.

### Parallel Opportunities

- T002 (Setup) can run in parallel with T001.
- T005, T006 (US1 Tests) can run in parallel with implementation tasks T007-T009.
- T012 (US2 Integration Test) can run in parallel with US1 implementation once the foundation is ready.
- All tasks marked [P] can run in parallel.

---

## Parallel Example: User Story 1

```bash
# Launch tests and implementation for User Story 1 together:
Task: "Create unit tests for SaveRaw logic in internal/clipboard/exporter_test.go"
Task: "Implement SaveRaw(path string, content models.ClipboardContent) error helper in internal/clipboard/exporter.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup.
2. Complete Phase 2: Foundational.
3. Complete Phase 3: User Story 1 (The core saving logic).
4. **STOP and VALIDATE**: Test User Story 1 independently using a browser copy and the `--save-raw` flag.

### Incremental Delivery

1. Foundation ready (Phase 2).
2. Add User Story 1 → Test independently → MVP!
3. Add User Story 2 → Test integration with stdout.
4. Final Polish and validation.

---

## Notes

- [P] tasks = different files, no dependencies.
- [Story] label maps task to specific user story for traceability.
- Each user story should be independently completable and testable.
- Verify tests fail before implementing logic.
- Commit after each task or logical group.
- Follow the project's silence-on-success and silence-on-empty policies.
