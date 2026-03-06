# md-paste Development Guidelines

Auto-generated from all feature plans. Last updated: 2026-03-06

## Active Technologies

- **Language**: Go 1.26+
- **CLI Framework**: `cobra`
- **Markdown Processing**: `html-to-markdown`
- **Testing**: `testify/assert`, `testify/require`

## Project Structure

```text
cmd/        # Main applications (entry points)
internal/   # Private application and library code
tests/      # Additional integration tests
```

## Commands

- `make setup`: Bootstrap the project (install deps)
- `make test`: Run unit tests
- `make test-integration`: Run integration tests (requires `MD_PASTE_E2E=1`)
- `make lint`: Run linters (`golangci-lint`)
- `make format`: Auto-format code (`gofumpt`)
- `make build`: Compile the application
- `make run`: Run the application locally

## Code Style

- Follow standard Go conventions.
- Format all code using `gofumpt`.
- Adhere to `golangci-lint` rules defined in `.golangci.yml`.

## Recent Changes

- `001-paste-markdown`: Initialized Go 1.26+ project with `cobra`, `html-to-markdown`, `testify`, automated releases via GoReleaser, and native macOS CGO clipboard integration.

<!-- MANUAL ADDITIONS START -->

## Architecture Decisions

### 1. Native macOS Clipboard via CGO
- **Context:** Need high-fidelity access to the macOS clipboard, specifically `public.html` (NSPasteboardTypeHTML), which standard Go libraries ignore.
- **Decision:** Implemented a direct CGO wrapper (`clipboard_darwin.go`) targeting `Foundation` and `AppKit` frameworks. Provided a stub (`clipboard_stub.go`) for non-darwin platforms to ensure cross-compilation success.
- **Rationale:** Native integration guarantees access to rich-text data copied from browsers and word processors.
- **Constraints:** Requires a `macos-latest` runner for GitHub Actions. Cross-platform testing requires conditional `//go:build darwin` tags.

### 2. Silence-on-Success CLI Design
- **Context:** The tool operates within a pipeline or as a background utility (copy/paste workflow).
- **Decision:** Implemented a strict silence policy where successful clipboard writes output nothing to stdout/stderr. Configured Cobra with `SilenceUsage: true` and `SilenceErrors: true`.
- **Rationale:** Aligns with the Unix philosophy ("no news is good news") and prevents terminal clutter.
- **Constraints:** Any necessary debug output must be routed to stderr via `slog` to keep the stdout pipe clean for data.

### 3. GoReleaser & Custom Homebrew Tap
- **Context:** Need a frictionless distribution mechanism for macOS users that avoids the strict notability requirements of the official `homebrew-core` repository.
- **Decision:** Configured a `.goreleaser.yaml` pipeline attached to GitHub Actions that automatically pushes a Ruby formula to a separate, custom `stn1slv/homebrew-tap` repository.
- **Rationale:** A custom tap (`brew install stn1slv/tap/md-paste`) provides the exact same UX as core Homebrew but gives full control over the release lifecycle and bypasses manual PR reviews from the Homebrew team.
- **Constraints:** Requires a dedicated `TAP_GITHUB_TOKEN` secret in the main repository with `repo` scope to allow the GitHub Action to cross-publish to the tap repository. Also requires using the safer array syntax in Ruby `system` calls (`system "#{bin}/md-paste", "--help"`) to avoid shell evaluation vulnerabilities.

## Known Issues & Gotchas

### ⚠️ Destructive Clipboard Integration Tests
- **Issue:** Integration tests mutating the system clipboard would permanently destroy a developer's local clipboard content if it contained non-restorable data (like images or complex HTML).
- **Root Cause:** The `t.Cleanup` function naively assumed it could restore state by rewriting plain text or clearing the clipboard, permanently losing rich media.
- **Prevention Rule:** Always check the initial state (`ContentTypeNone` or missing `PlainText`) and invoke `t.Skip` if the clipboard cannot be safely and perfectly restored. Gate all system-mutating tests behind an explicit environment variable (e.g., `MD_PASTE_E2E=1`).

### ⚠️ CGO Pointers & Empty String Types
- **Issue:** The Objective-C code returned a non-nil `cHTML` pointer even when the underlying string was empty, causing the Go application to prioritize empty HTML over valid fallback plain text.
- **Root Cause:** macOS `NSPasteboard` can populate types without meaningful payloads.
- **Prevention Rule:** When reading from CGO pointers, explicitly convert to `C.GoString` and check `if str != ""` before mutating state variables or assuming priority content exists.

### ⚠️ Cobra Stream Redirection in Tests
- **Issue:** Using `fmt.Println` to output the `--stdout` result broke Cobra's stream redirection, making CLI testing extremely difficult and potentially breaking embedded use cases.
- **Root Cause:** Hardcoding global standard streams rather than relying on the command context.
- **Prevention Rule:** Always use `fmt.Fprintln(cmd.OutOrStdout(), ...)` or `cmd.Println(...)` within Cobra `RunE` functions to respect internal stream overrides.

<!-- MANUAL ADDITIONS END -->
