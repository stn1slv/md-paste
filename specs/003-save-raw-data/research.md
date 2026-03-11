# Research: Save raw clipboard data to file

## Decisions

### Decision 1: File Writing Implementation
- **Choice**: Use `os.WriteFile` for atomic-style writing of the raw content.
- **Rationale**: `os.WriteFile` handles opening, writing, and closing the file in one call. It truncates the file if it exists, fulfilling the "File Overwrite" assumption.
- **Alternatives Considered**: `os.Create` followed by `io.Copy` or `f.Write`. These are more verbose and unnecessary for the relatively small payloads (HTML/Text) expected from the clipboard.

### Decision 2: Directory Path Validation
- **Choice**: Use `os.Stat` to check if the provided path is a directory before attempting to write.
- **Rationale**: Explicitly handles the clarification from the specification: "Return a descriptive error if the path is a directory". While `os.WriteFile` would fail if the path is a directory, the error message from the OS might be less descriptive than a custom one.
- **Alternatives Considered**: Relying solely on `os.WriteFile` error. Rejected because we want to provide a specific, user-friendly error message as per FR-005.

### Decision 3: "Raw" Data Definition
- **Choice**: Save exactly what is retrieved from the `clipboard.Read()` function without any modification.
- **Rationale**: Ensures "Source Integrity" as per the specification. If `RawHTML` is populated, save that; otherwise, save `PlainText`.
- **Alternatives Considered**: Attempting to format or "clean" the HTML before saving. Rejected as it contradicts the "raw" requirement.

## Best Practices

- **Go File Permissions**: Use `0644` (readable by all, writable by owner) as the default permission for the created file, which is standard for user-generated documents.
- **Error Wrapping**: Follow the project's constitution (Principle V) by wrapping file system errors with context (e.g., `fmt.Errorf("failed to save raw data to %s: %w", path, err)`).
