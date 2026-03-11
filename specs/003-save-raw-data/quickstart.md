# Quickstart: Save raw clipboard data to file

## Overview

`md-paste` now supports an optional `--save-raw` (`-r`) flag to persist unconverted clipboard data (HTML or Plain Text) to a file on the local filesystem. This is useful for debugging and archiving source payloads.

## Basic Usage

1.  **Copy some rich text** from a browser (e.g., this document).
2.  **Run the command** to save the raw HTML:
    ```bash
    md-paste --save-raw raw.html
    ```
3.  **Inspect the file**:
    ```bash
    cat raw.html
    ```

## Advanced Usage

Combine with `--stdout` for a single-command export pipeline:
```bash
md-paste -r original.html -s > result.md
```

## Error Handling

If you provide a directory path instead of a file path, the command will fail with a descriptive error:
```bash
md-paste -r /tmp
# Error: '/tmp' is a directory
```

If the clipboard is empty, no file will be created, and the command will exit silently (following the project's silence-on-empty policy).
