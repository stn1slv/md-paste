# Data Model: Paste Markdown CLI (001-paste-markdown)

## ClipboardContent (macOS)
**Attributes**:
- `RawHTML (string)`: The raw HTML string fetched from the `NSPasteboard` (`public.html`).
- `PlainText (string)`: Fallback UTF-8 string if HTML is unavailable.
- `ContentType (enum)`: `HTML`, `PlainText`, `None`.

**Validation Rules**:
- If `RawHTML` is not empty, it takes priority over `PlainText`.
- If both are empty, it's considered an empty clipboard scenario.

## MarkdownDocument
**Attributes**:
- `Content (string)`: The final Markdown string after conversion.
- `SourceType (enum)`: Indicates if it was derived from `HTML` or `PlainText`.

**Validation Rules**:
- Must be a valid UTF-8 string.
- Should follow GFM/CommonMark syntax.
