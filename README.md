# md-paste

A fast, native macOS CLI application that converts rich text (HTML) from your clipboard into standard Markdown, automatically saving it back to your clipboard so you can paste it directly into your favorite Markdown editor.

## Why md-paste?
When copying from a browser or a word processor, the clipboard stores rich text. Pasting this into a code editor or a markdown file usually strips formatting or pastes raw HTML. `md-paste` seamlessly bridges this gap by intercepting the clipboard and transforming it to GitHub Flavored Markdown (GFM).

## Features
- **Native Integration**: Uses `NSPasteboard` via CGO for high-fidelity clipboard access.
- **HTML Priority**: Automatically detects HTML and converts it. Falls back to plain text if needed.
- **Pipe-Friendly**: Unix philosophy support with an optional `--stdout` flag.
- **Menu Bar / Tray App (macOS & Windows)**: Optional resident app to convert the clipboard with a single click, with a "Launch at login" toggle.
- **Configurable Global Shortcut**: Trigger a conversion from any app with a customizable system-wide hotkey (default `Cmd+Opt+M` on macOS, `Ctrl+Alt+M` on Windows).

## Installation

### Via Homebrew (Recommended)
```bash
brew install stn1slv/tap/md-paste
```

### Menu Bar App (macOS)
Install the menu bar app as a Homebrew Cask:
```bash
brew install --cask stn1slv/tap/md-paste
```
The app is ad-hoc signed and not notarized, so on first launch macOS Gatekeeper
may block it. If that happens, clear the quarantine attribute:
```bash
xattr -dr com.apple.quarantine "/Applications/md-paste.app"
```

### Windows
Install via WinGet:
```powershell
winget install stn1slv.md-paste
```
This installs two commands: `md-paste` (the console CLI) and `md-paste-tray` (the
system tray app). You can also download the Windows zip from the
[Releases](https://github.com/stn1slv/md-paste/releases) page, which contains both
`md-paste.exe` and `md-paste-tray.exe`.

### Build from source
#### Prerequisites
- macOS (requires native Cocoa/AppKit libraries) or Windows
- Go 1.26+

#### Build steps
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

- **Run as a menu bar app (macOS)**:
  ```bash
  md-paste menubar
  ```
  Adds an icon to the menu bar. Click "Convert clipboard to Markdown" to convert
  the clipboard in place (the icon briefly shows a checkmark on success), or press
  the global shortcut (default `Cmd+Opt+M`) from any app. Use the "Launch at login"
  toggle to start the app automatically after you log in.

- **Run as a system tray app (Windows)**:
  Launch `md-paste-tray.exe` (no console window appears). Right-click the tray
  icon and choose "Convert clipboard to Markdown" to convert the clipboard in
  place (the icon briefly turns into a checkmark on success), or press the global
  shortcut (default `Ctrl+Alt+M`) from any app. Use the "Launch at login" toggle to
  start it automatically at sign-in. The console `md-paste.exe` still works as the
  CLI, including `md-paste --stdout`.

### Global shortcut and configuration

The resident app registers a system-wide keyboard shortcut that runs a conversion
from anywhere. Settings live in a YAML config file that the app creates on first
run:

- macOS: `~/Library/Application Support/md-paste/config.yaml`
- Windows: `%AppData%\md-paste\config.yaml`

```yaml
hotkey: cmd+opt+m       # macOS default; use e.g. "ctrl+alt+m" on Windows
launch_at_login: false
```

The `hotkey` value is modifiers plus a key joined by `+`. Recognized modifiers are
`cmd`, `ctrl`, `shift`, `opt`/`alt` (and `win` on Windows); the key can be a
letter, a digit, `space`, or `f1`-`f12`. Pick a combination with at least one
modifier.

To change the shortcut, use the menu items:

- **Edit shortcut...** opens the config file in your default editor.
- **Reload config** re-reads the file and re-registers the shortcut live, with no
  restart.

The "Launch at login" toggle and the config's `launch_at_login` field stay in sync;
the config file is the source of truth.

Notes:

- macOS registration uses the Carbon hot key API and needs **no** Accessibility
  permission.
- On Windows, if a combination is already claimed by another app it cannot be
  registered; choose a different one in the config and reload.

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
