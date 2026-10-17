# md-paste

md-paste converts rich text (HTML) on your clipboard to Markdown and writes the result back to the clipboard, so you can paste it straight into a Markdown editor. It is a native CLI for macOS and Windows, and it can also run as a menu bar app (macOS) or a system tray app (Windows).

## Why md-paste?
When you copy from a browser or a word processor, the clipboard holds rich text. Pasting it into a code editor or a Markdown file usually drops the formatting or pastes raw HTML. md-paste reads the clipboard and converts its content to GitHub Flavored Markdown (GFM).

## Features
- Native clipboard access: `NSPasteboard` via CGO on macOS, the Win32 clipboard API on Windows.
- Converts HTML when the clipboard has it and falls back to plain text otherwise.
- Works in pipelines: `--stdout` prints the Markdown instead of writing it to the clipboard.
- Optional menu bar / tray app (macOS and Windows) that converts the clipboard with one click and has a "Launch at login" toggle.
- Configurable global shortcut that starts a conversion from any app (default `Ctrl+Cmd+Opt+M` on macOS, `Ctrl+Alt+M` on Windows).

## Installation

### Homebrew (recommended)
```bash
brew install stn1slv/tap/md-paste
```

### Menu bar app (macOS)
Install the menu bar app as a Homebrew Cask:
```bash
brew install --cask stn1slv/tap/md-paste
```
The app is ad-hoc signed and not notarized, so macOS Gatekeeper may block it on first launch. If it does, remove the quarantine attribute:
```bash
xattr -dr com.apple.quarantine "/Applications/md-paste.app"
```

### Windows
Install with WinGet:
```powershell
winget install stn1slv.md-paste
```
This installs two commands: `md-paste` (the console CLI) and `md-paste-tray` (the system tray app). The Windows zip on the [Releases](https://github.com/stn1slv/md-paste/releases) page contains the same two programs, `md-paste.exe` and `md-paste-tray.exe`.

### Build from source
#### Prerequisites
- macOS (requires native Cocoa/AppKit libraries) or Windows
- Go 1.26+
- `gofumpt` and `golangci-lint` (`make setup` installs them with Homebrew; on Windows install them yourself)

#### Build steps
```bash
git clone https://github.com/stn1slv/md-paste.git
cd md-paste
make setup
make build
```
The binary is written to `./bin/md-paste`.

## Usage

1. Copy rich text from your browser, Google Docs, MS Word, etc. (Cmd+C).
2. Run `md-paste`.
3. Paste into your Markdown editor (Cmd+V).

### Options

- Convert and write the result back to the clipboard (default):
  ```bash
  md-paste
  ```
  It prints nothing when it succeeds, so it fits into scripts and shortcuts.

- Convert and print to the terminal:
  ```bash
  md-paste --stdout
  # or
  md-paste -s
  ```

- Pipe the result to other commands:
  ```bash
  md-paste -s | grep "TODO"
  ```

- Run as a menu bar app (macOS):
  ```bash
  md-paste menubar
  ```
  This adds an icon to the menu bar. Click "Convert clipboard to Markdown" to convert the clipboard in place (the icon shows a checkmark for a moment on success), or press the global shortcut (default `Ctrl+Cmd+Opt+M`) from any app. Turn on "Launch at login" to start the app after you log in.

- Run as a system tray app (Windows):
  Launch `md-paste-tray.exe`; no console window opens. Right-click the tray icon and choose "Convert clipboard to Markdown" to convert the clipboard in place (the icon turns into a checkmark for a moment on success), or press the global shortcut (default `Ctrl+Alt+M`) from any app. Turn on "Launch at login" to start it at sign-in. The console `md-paste.exe` still works as the CLI, including `md-paste --stdout`.

### Global shortcut and configuration

The resident app registers a system-wide keyboard shortcut that runs a conversion from any app. Its settings are in a YAML config file that the app creates on first run:

- macOS: `~/Library/Application Support/md-paste/config.yaml`
- Windows: `%AppData%\md-paste\config.yaml`

```yaml
hotkey: ctrl+cmd+opt+m  # macOS default; use e.g. "ctrl+alt+m" on Windows
launch_at_login: false
```

The `hotkey` value is one or more modifiers and a key, joined by `+`. The modifiers are `cmd`, `ctrl`, `shift`, `opt`/`alt` (and `win` on Windows). The key can be a letter, a digit, `space`, or `f1`-`f12`. At least one modifier is required.

To change the shortcut, use these menu items:

- **Edit shortcut...** opens the config file in your default editor.
- **Reload config** reads the file again and registers the new shortcut without a restart.

The "Launch at login" toggle and the `launch_at_login` field stay in sync. If they disagree, the config file wins.

Notes:

- On macOS the shortcut uses the Carbon hot key API, which does not need Accessibility permission.
- On Windows a combination that another app already uses cannot be registered. Pick a different one in the config and reload.

## Development
See the [Constitution](.specify/memory/constitution.md) for core principles.

```bash
make setup            # Install dependencies
make test             # Run unit tests
make test-integration # Run unit + E2E clipboard tests (modifies the system clipboard)
make lint             # Run golangci-lint
make format           # Run gofumpt
make bundle           # Build the macOS menu bar .app bundle and zip (in ./dist)
```
