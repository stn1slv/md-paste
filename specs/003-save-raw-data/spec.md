# Feature Specification: Save raw clipboard data to file

**Feature Branch**: `003-save-raw-data`  
**Created**: 2026-03-11  
**Status**: Draft  
**Input**: User description: "I want to introduce optional flag for saving to the file raw data from clipboard"

## Clarifications

### Session 2026-03-11
- Q: How should the system handle a directory path provided to `--save-raw`? → A: Return a descriptive error (e.g., "Error: './path' is a directory") and exit.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Debugging Rich Text (Priority: P1)

As a developer or power user, I want to see the exact HTML that `md-paste` is receiving from the macOS clipboard so I can debug conversion issues or understand how different browsers format their copy payloads.

**Why this priority**: This is the primary driver for the feature - visibility into the raw source data.

**Independent Test**: Can be fully tested by copying rich text from a browser and running `md-paste --save-raw raw.html`, then comparing `raw.html` with the expected HTML.

**Acceptance Scenarios**:

1. **Given** the clipboard contains rich text (HTML), **When** I run `md-paste --save-raw raw.html`, **Then** a file named `raw.html` is created containing the raw HTML content, and the converted Markdown is still written to the clipboard.
2. **Given** the clipboard contains plain text only, **When** I run `md-paste --save-raw raw.txt`, **Then** a file named `raw.txt` is created containing the plain text content.

---

### User Story 2 - Integrated Export Pipeline (Priority: P2)

As a user, I want to save the raw source data while simultaneously outputting the converted Markdown to standard output so I can archive the original and process the result in a single command.

**Why this priority**: Improves utility in scripts and complex workflows.

**Independent Test**: Can be tested by running `md-paste --save-raw raw.html --stdout > result.md` and verifying both files.

**Acceptance Scenarios**:

1. **Given** valid clipboard content, **When** I run `md-paste --save-raw raw.html --stdout`, **Then** the raw data is saved to `raw.html` AND the Markdown is printed to stdout.

---

### Edge Cases

- **Empty Clipboard**: Following the project's "Silence-on-Empty" policy, if the clipboard is empty, no file should be created, and the tool should exit silently with a success code.
- **Pre-existing File**: The tool should overwrite the file if it already exists, matching standard CLI behavior for output flags.
- **Unwritable Path**: If the path is invalid (e.g., directory doesn't exist) or permission is denied, the tool must report a clear error and exit with a non-zero code.
- **No HTML/Plain Text**: If the clipboard contains data that is neither HTML nor Plain Text (e.g., just an image), it should be treated as "Empty" (since `md-paste` focuses on text/html).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide a CLI flag `--save-raw` (short version `-r`) that accepts a file path as an argument.
- **FR-002**: System MUST save the original, unconverted content from the clipboard to the specified file path.
- **FR-003**: System MUST prioritize HTML content for the raw save if available; if HTML is absent, it MUST use Plain Text.
- **FR-004**: If the clipboard is empty (no HTML and no Plain Text), the system MUST NOT create the file and MUST exit silently.
- **FR-005**: System MUST exit with a descriptive error if the file path is invalid, is a directory, or the file cannot be written due to permissions.
- **FR-006**: Saving raw data MUST be an additional action; it MUST NOT prevent the default behavior of converting content to Markdown and writing it to the clipboard or stdout.

### Key Entities

- **Raw Content**: The byte-for-byte representation of either `public.html` or `public.utf8-plain-text` from the macOS pasteboard.
- **Target File**: A file on the local filesystem where the raw content is persisted.

## Assumptions

- **File Overwrite**: The user expects the target file to be overwritten if it already exists, consistent with standard CLI output flags.
- **HTML Priority**: When both HTML and Plain Text are on the clipboard, the HTML version is more valuable for "raw" data analysis/debugging and should be prioritized.
- **Source Integrity**: The "raw" data refers to the content exactly as retrieved by the system's clipboard API, before any internal processing or sanitization.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: The saved file contains exactly the same content retrieved from the clipboard (HTML or Plain Text).
- **SC-002**: Zero output to stdout/stderr on success (unless `--stdout` is used, in which case only Markdown goes to stdout).
- **SC-003**: Existing integration tests for Markdown conversion continue to pass without modification.
