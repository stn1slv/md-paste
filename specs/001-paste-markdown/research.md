# Research: Paste Markdown CLI (001-paste-markdown)

## Decision: HTML to Markdown Library
**Choice**: `github.com/JohannesKaufmann/html-to-markdown`
**Rationale**: This is the most popular and robust Go library for converting HTML to Markdown. It is highly configurable and aims for compatibility with common Markdown flavors (GFM, CommonMark), satisfying SC-001.
**Alternatives considered**:
- `github.com/jaytaylor/html2text`: More focused on plain text than structured Markdown.
- Writing a custom parser: Unnecessary complexity; violates "Simplicity First".

## Decision: macOS Clipboard Integration
**Choice**: `github.com/atotto/clipboard` for general text, but for `NSPasteboard` with HTML support, we need a specialized approach. We will use a CGO wrapper or a library like `github.com/progrium/macdriver` or similar, but for a simple CLI, a direct CGO implementation or calling `pbpaste`/`pbcopy` might be considered. However, the spec EXPLICITLY requires `NSPasteboard` using CGO (Non-Functional Requirement 2).
**Rationale**: Native CGO with `Foundation` and `AppKit` frameworks ensures we can access `public.html` (or `NSPasteboardTypeHTML`) specifically, which standard Go clipboard libraries often skip.
**Alternatives considered**:
- `pbpaste -Prefer html`: Reliable but introduces a shell dependency; spec requires direct integration.

## Decision: CLI Framework
**Choice**: `github.com/spf13/cobra`
**Rationale**: Mandated by the project constitution for "Robust CLI Interface".
**Alternatives considered**: None (Constitution mandate).
