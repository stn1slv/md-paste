# Feature Specification: Paste Markdown CLI

**Feature Branch**: `001-paste-markdown`  
**Created**: 2026-03-06  
**Status**: Draft  
**Input**: User description: "# Application Requirements: Paste Markdown CLI ## Overview A Command Line Interface (CLI) application written in Golang that reads rich text (HTML) copied from a browser or document, converts it to Markdown, and saves the converted result back to the clipboard. ## Functional Requirements 1. **Clipboard Reading**: The application must read data from the system clipboard. 2. **HTML Priority**: The application must attempt to extract rich text (HTML format) from the clipboard. If HTML is not present, it should fall back to reading plain text. 3. **Markdown Conversion**: The application must convert the retrieved HTML string into standard Markdown syntax. The conversion should match the behavior of popular tools like `turndown` or `to-markdown`. 4. **Clipboard Writing**: As the default behavior, the application must automatically write the converted Markdown string back to the system clipboard so the user can immediately paste it. 5. **Stdout Output (Optional)**: The application must support an optional command-line flag to print the converted Markdown text directly to stdout *instead* of writing it back to the clipboard. ## Non-Functional Requirements 1. **Language**: The application must be written in Go (Golang) and pinned to the latest stable version (e.g., 1.26+). 2. **Operating System**: The application targets macOS and must integrate directly with the macOS clipboard (e.g., via `NSPasteboard` using CGO). 3. **Execution**: The application will run as a standalone CLI executable. 4. **Project Structure**: The application must follow standard Go project layouts (`cmd/md-paste/`, `internal/`). 5. **Tooling & Code Quality**: The repository must abide by the provided Go Application Blueprint: - **Makefile**: Include `setup`, `test`, `lint`, `format`, `build`, and `upgrade-deps`. - **Formatting**: Must use `gofumpt` for formatting. - **Linting**: Must use `golangci-lint` (with `revive`, `gocritic`, `gosec`, `prealloc`, `govet`, `staticcheck` enabled). - **Testing**: Must use standard `testing` package with Table-Driven Tests and `testify/assert` or `testify/require`. - **Logging**: Use standard library `log/slog` for any logging output."

## Clarifications

### Session 2026-03-06
- Q: Success feedback behavior? → A: Silence (Unix style)
- Q: Behavior for non-text/HTML data on clipboard? → A: Silence & Exit (Treat as empty)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Convert Clipboard HTML to Markdown (Priority: P1)

As a technical writer, I want to copy a rich-text snippet from my browser and have it automatically converted to Markdown on my clipboard, so I can paste it into my Markdown editor without manual reformatting.

**Why this priority**: This is the core functionality and primary use case of the application.

**Independent Test**: Copy a snippet from a website (e.g., a header and a list), run `md-paste`, and then paste into a text editor. The output should be valid Markdown.

**Acceptance Scenarios**:

1. **Given** rich text (HTML) is on the system clipboard, **When** `md-paste` is executed without flags, **Then** the clipboard content is replaced with the equivalent Markdown string and NO output is printed to the terminal.
2. **Given** plain text is on the system clipboard, **When** `md-paste` is executed, **Then** the plain text is treated as Markdown (no-op conversion or basic escaping) and written back to the clipboard.
3. **Given** non-text data (e.g., an image) is on the system clipboard, **When** `md-paste` is executed, **Then** the application exits silently with no changes to the clipboard and no output.

---

### User Story 2 - Output Markdown to Stdout (Priority: P2)

As a power user, I want to convert clipboard content and pipe the Markdown output into another tool or a file, so I can integrate it into my automated workflows.

**Why this priority**: Supports the Unix philosophy of tool composability as defined in the project constitution.

**Independent Test**: Run `md-paste --stdout` (or similar flag) and verify that the Markdown is printed to the terminal and NOT written back to the clipboard.

**Acceptance Scenarios**:

1. **Given** rich text is on the clipboard, **When** `md-paste --stdout` is executed, **Then** the converted Markdown is printed to stdout and the clipboard remains unchanged.

---

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST read the system clipboard using macOS native APIs (`NSPasteboard`).
- **FR-002**: System MUST prioritize HTML content (`public.html` or equivalent) if available on the clipboard.
- **FR-003**: System MUST fall back to UTF-8 plain text if no HTML content is found.
- **FR-003.1**: System MUST exit silently and perform no write if neither HTML nor plain text is found on the clipboard.
- **FR-004**: System MUST convert HTML to standard CommonMark/GFM compatible Markdown.
- **FR-005**: System MUST write the converted Markdown back to the clipboard as the default operation.
- **FR-005.1**: System MUST NOT output any text to stdout or stderr on successful clipboard write (Silence-on-Success).
- **FR-006**: System MUST provide a command-line flag (e.g., `--stdout` or `-s`) to redirect output to the terminal instead of the clipboard.
- **FR-007**: System MUST provide a help flag (`--help`) describing usage and flags.

### Key Entities

- **Clipboard Content**: The raw data (HTML or Text) retrieved from the macOS pasteboard.
- **Markdown Document**: The structured text result of the conversion process.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Conversion of standard HTML elements (headers, lists, links, bold/italic) matches the output of `turndown` with 100% parity.
- **SC-002**: Execution time from command invocation to clipboard update is under 200ms for payloads up to 100KB.
- **SC-003**: The application successfully handles and converts clipboard data from Safari, Chrome, and Microsoft Word.
- **SC-004**: 100% of user-triggered errors (e.g., internal system failure) are reported to stderr with a non-zero exit code.
