# Quickstart: Table Support

This guide demonstrates how to use the new table support features in `md-paste`.

## Prerequisites

- Ensure you have `md-paste` installed and running.
- You have a document open in Word, Excel, Confluence, or a PDF viewer.

## Scenario 1: Copying from Word or Excel

1. Open a document containing a table with headers and data.
2. Highlight the table and copy it to your clipboard (`Cmd+C`).
3. Focus your Markdown editor or terminal.
4. Run `md-paste`.
5. **Result**: The table is instantly inserted as a perfectly formatted GitHub Flavored Markdown (GFM) table, preserving your header row.

## Scenario 2: Copying from Confluence

1. In Confluence, highlight a table that includes aligned columns and status macros (e.g., "IN PROGRESS").
2. Copy the table.
3. Run `md-paste` in your editor.
4. **Result**: The output is a GFM table. The alignment markers (like `---:`) are set correctly, and the Confluence status macros are cleanly converted to plain text.

## Scenario 3: Copying from a PDF

1. Open a PDF containing tabular data.
2. Select the text block that forms the table and copy it.
3. Run `md-paste`.
4. **Result**: `md-paste` analyzes the spacing and translates the text block into a structured GFM table, saving you the effort of manually adding pipes `|` and dashes `-`.
