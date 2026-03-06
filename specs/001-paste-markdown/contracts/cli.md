# CLI Contract: Paste Markdown CLI (001-paste-markdown)

## Command: `md-paste` (default)
**Purpose**: Reads the clipboard, converts it to Markdown, and writes it back to the clipboard.

### Flags
| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--stdout` | `-s` | `bool` | `false` | If true, the converted Markdown is printed to stdout and NOT written to the clipboard. |
| `--help` | `-h` | `bool` | `false` | Displays help information. |

### Outputs
- **Silence (Default)**: If no flags are provided and conversion is successful, the command exits with code 0 and NO output.
- **stdout**: If `--stdout` is provided, the Markdown content is printed to the terminal.
- **stderr**: Real internal failures (e.g., system or conversion errors) are printed to stderr with a non-zero exit code. Note: An empty clipboard or non-text clipboard data is considered a silent success (code 0) and produces no error.

### Exit Codes
| Code | Meaning |
|------|---------|
| `0` | Success (even if no conversion was possible but it wasn't a "failure") |
| `1` | General error / Internal system failure |
| `126` | Command invoked cannot execute |
| `127` | Command not found |
