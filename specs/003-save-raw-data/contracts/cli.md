# CLI Contract: Save raw clipboard data to file

## Commands

### `md-paste` (Main)

The `md-paste` command is updated to support optional raw data saving.

#### Flags

| Name | Type | Short | Description | Default |
|------|------|-------|-------------|---------|
| `--save-raw` | `string` | `-r` | File path where raw clipboard data will be saved. | "" |

#### Usage Examples

```bash
# Save raw HTML to debug.html and write Markdown back to clipboard
md-paste --save-raw debug.html

# Save raw data and print Markdown to stdout
md-paste -r source.txt -s

# If directory exists, should fail with error
mkdir results
md-paste -r results
# Output: Error: failed to save raw content: 'results' is a directory
```

#### Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success (including empty clipboard scenarios). |
| `1` | Failure (conversion error, file system error, path is directory). |

#### Environment Variables

- None specifically added for this feature.
