# Research: Table Support

## 1. Parsing HTML Tables (Word, Excel, Confluence)

**Decision**: Use `golang.org/x/net/html` or enhance the existing `html-to-markdown` library to handle `<table>`, `<tr>`, `<th>`, and `<td>` tags. We will use a DOM traversal approach to build an intermediate table representation before outputting GFM.
**Rationale**: `x/net/html` is the standard library for HTML parsing in Go and is robust enough to handle malformed HTML often produced by office applications. Using an intermediate representation allows us to easily handle row/col spanning (flattening) and calculate column alignments before rendering to Markdown.
**Alternatives considered**: Using `PuerkitoBio/goquery` (adds an external dependency, which might be overkill if we just need table traversal), regex (brittle and error-prone for HTML).

## 2. Table Column Alignment

**Decision**: Parse `style` attribute for `text-align` (e.g., `text-align: right;`) or standard attributes like `align="right"` on `<th>` and `<td>` elements to determine column alignment for the GFM separator row.
**Rationale**: Confluence and Excel often export alignment in inline styles or attributes.
**Alternatives considered**: Ignoring alignment (fails user requirements), guessing based on content (unreliable).

## 3. Merged Cells Strategy

**Decision**: Flatten merged cells by repeating the content of the top-left cell into all spanned rows and columns.
**Rationale**: GFM does not support `rowspan` or `colspan`. Flattening preserves the table's structural grid in Markdown, ensuring it renders correctly without breaking column counts.
**Alternatives considered**: Outputting raw HTML `<table>` (violates user request for GFM), leaving spanned cells blank (can cause data misalignment visually).

## 4. PDF Plain Text Reconstruction

**Decision**: Implement a layout-aware heuristic parser that splits lines on 2 or more contiguous whitespace characters (`\s{2,}`) or tabs (`\t`) to detect column boundaries.
**Rationale**: PDF text copied to the clipboard usually separates visually distant columns with multiple spaces or tabs.
**Alternatives considered**: NLP-based tabular extraction (too heavy/slow), strict CSV/TSV parsing (fails on generic PDF text).

## 5. Confluence Macros

**Decision**: Identify Confluence-specific HTML structures (e.g., `<span class="confluence-jim-macro">`) and extract only the inner text (e.g., the status text), stripping the surrounding formatting.
**Rationale**: Maps directly to the requirement to strip them to plain text representation.
**Alternatives considered**: Mapping to specific Markdown badges (too complex and flavor-specific).
