# Data Model: Save raw clipboard data to file

## Entities

### `ClipboardContent` (Existing)
*Modified interpretation in this feature context.*

| Field | Type | Description |
|-------|------|-------------|
| `RawHTML` | `string` | The raw HTML retrieved from `public.html`. Prioritized for saving. |
| `PlainText` | `string` | Fallback if `RawHTML` is empty. |
| `ContentType` | `models.ContentType` | Used to determine if clipboard is empty. |

### `TargetFile` (New)
*A physical entity on the filesystem.*

| Attribute | Validation | Description |
|-----------|------------|-------------|
| `Path` | MUST be a valid file path | Provided via `--save-raw`. |
| `Permissions` | `0644` | Default permissions for the saved file. |
| `Exists` | Overwrite | If file exists, it will be truncated. |
| `IsDirectory` | MUST NOT be a directory | `os.Stat` must return `IsDir() == false`. |

## Relationships

- **ClipboardContent → TargetFile**: A one-way "save" operation. The `ClipboardContent` is mapped to the `TargetFile` based on content priority (HTML > PlainText).
- **TargetFile → Filesystem**: Interaction through `os.WriteFile` and `os.Stat`.

## State Transitions

1. **Retrieved**: Content is in memory after `clipboard.Read()`.
2. **Validated**: Path is checked via `os.Stat` to ensure it's not a directory.
3. **Persisted**: Content is written to the filesystem via `os.WriteFile`.
4. **Finalized**: CLI exits silently (on success) or with an error (on failure).
