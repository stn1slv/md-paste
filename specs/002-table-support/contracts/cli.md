# CLI Contracts: Table Support

This document describes any changes or additions to the CLI interface for `md-paste` introduced by the table support feature.

## CLI Behavior Changes

No new explicit CLI flags are added in this iteration. The tool's default behavior is enhanced to automatically detect and convert tables.

### Implicit Behavior (Standard Execution)

```bash
md-paste
```

**Input**: User executes `md-paste` while a table (from Word, Excel, Confluence, or PDF) is in the system clipboard.
**Behavior**: 
1. The tool checks for `public.html`. If found, it parses it as an HTML table.
2. If `public.html` is not found, it checks `public.utf8-plain-text`. If the text contains tabular heuristics (e.g., multiple spaces separating data on multiple lines), it parses it as a text table.
3. Converts the parsed table to GFM format.
**Output**: The resulting GFM string is pasted to standard output (if configured) or typed into the active window (based on base application behavior).

### Error/Fallback Handling

If a table parsing fails or the input is determined not to be a table despite initial heuristics, `md-paste` gracefully falls back to its standard plain text or simple markdown formatting behavior without crashing.
