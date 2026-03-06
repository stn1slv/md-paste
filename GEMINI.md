# md-paste Development Guidelines

Auto-generated from all feature plans. Last updated: 2026-03-06

## Active Technologies

- Go 1.26+ + `cobra`, `html-to-markdown`, `testify/assert`, `testify/require` (001-paste-markdown)

## Project Structure

```text
cmd/
internal/
tests/
```

## Commands

# Add commands for Go 1.26+

## Code Style

Go 1.26+: Follow standard conventions

## Recent Changes

- 001-paste-markdown: Added Go 1.26+ + `cobra`, `html-to-markdown`, `testify/assert`, `testify/require`

<!-- MANUAL ADDITIONS START -->

## Architecture Decisions

### Native macOS Clipboard via CGO
**Context:** Need high-fidelity access to macOS clipboard, specifically `public.html` (NSPasteboardTypeHTML), which standard Go libraries ignore.
**Decision:** Implemented a direct CGO wrapper (`clipboard_darwin.go`) targeting `Foundation` and `AppKit` frameworks. Provided a `clipboard_stub.go` for non-darwin platforms to ensure cross-compilation success.
**Rationale:** Native integration guarantees access to rich-text data copied from browsers and word processors.
**Constraints:** Requires `macos-latest` runner for GitHub Actions. Cross-platform testing requires conditional `//go:build darwin` tags.

### Silence-on-Success CLI Design
**Context:** The tool operates within a pipeline or as a background utility (copy/paste workflow).
**Decision:** Implemented a strict silence policy where successful clipboard writes output nothing to stdout/stderr.
**Rationale:** Aligns with Unix philosophy ("no news is good news") and prevents terminal clutter.
**Constraints:** Any necessary debug output must be routed to stderr via `slog` to keep the stdout pipe clean for data.

## Known Issues & Gotchas

### ⚠️ Destructive Clipboard Integration Tests
**Issue:** Integration tests mutating the system clipboard would permanently destroy a developer's local clipboard content if it contained non-restorable data (like images or complex HTML).
**Root Cause:** The `t.Cleanup` function naively assumed it could restore state by rewriting plain text or clearing the clipboard, permanently losing rich media.
**Prevention Rule:** Always check the initial state (`ContentTypeNone` or missing `PlainText`) and invoke `t.Skip` if the clipboard cannot be safely and perfectly restored. Gate all system-mutating tests behind an explicit environment variable (e.g., `MD_PASTE_E2E=1`).

### ⚠️ CGO Pointers & Empty String Types
**Issue:** The objective-C code returned a non-nil `cHTML` pointer even when the underlying string was empty, causing the Go application to prioritize empty HTML over valid fallback plain text.
**Root Cause:** macOS `NSPasteboard` can populate types without meaningful payloads.
**Prevention Rule:** When reading from CGO pointers, explicitly convert to `C.GoString` and check `if str != ""` before mutating state variables or assuming priority content exists.

### ⚠️ Cobra Stream Redirection in Tests
**Issue:** Using `fmt.Println` to output the `--stdout` result broke Cobra's stream redirection, making CLI testing extremely difficult and potentially breaking embedded use cases.
**Root Cause:** Hardcoding global standard streams rather than relying on the command context.
**Prevention Rule:** Always use `fmt.Fprintln(cmd.OutOrStdout(), ...)` or `cmd.Println(...)` within Cobra `RunE` functions to respect internal stream overrides.

<!-- MANUAL ADDITIONS END -->
