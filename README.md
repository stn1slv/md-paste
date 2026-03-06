# md-paste

A fast, native macOS CLI application that converts rich text (HTML) from your clipboard into standard Markdown, automatically saving it back to your clipboard so you can paste it directly into your favorite Markdown editor.

## Why md-paste?
When copying from a browser or a word processor, the clipboard stores rich text. Pasting this into a code editor or a markdown file usually strips formatting or pastes raw HTML. `md-paste` seamlessly bridges this gap by intercepting the clipboard and transforming it to GitHub Flavored Markdown (GFM).

## Features
- **Native Integration**: Uses `NSPasteboard` via CGO for high-fidelity clipboard access.
- **HTML Priority**: Automatically detects HTML and converts it. Falls back to plain text if needed.
- **Pipe-Friendly**: Unix philosophy support with an optional `--stdout` flag.

## Installation

### Prerequisites
- macOS (requires native Cocoa/AppKit libraries)
- Go 1.26+

### Build from source
```bash
git clone https://github.com/stn1slv/md-paste.git
cd md-paste
make setup
make build
```
The binary will be available in `./bin/md-paste`.

## Usage

1. Copy rich text from your browser, Google Docs, MS Word, etc. (Cmd+C).
2. Run `md-paste`.
3. Paste directly into your Markdown editor (Cmd+V).

### Options

- **Convert and save to clipboard (Default)**:
  ```bash
  md-paste
  ```
  *(Operates silently on success for optimal workflow integration.)*

- **Convert and print to terminal**:
  ```bash
  md-paste --stdout
  # or
  md-paste -s
  ```

- **Pipe to other commands**:
  ```bash
  md-paste -s | grep "TODO"
  ```

## Development
See the [Constitution](.specify/memory/constitution.md) for core principles.

```bash
make setup  # Install dependencies
make test   # Run unit and integration tests
make lint   # Run golangci-lint
make format # Run gofumpt
```
