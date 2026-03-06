# Quickstart: Paste Markdown CLI (001-paste-markdown)

## Local Setup
1. Clone the repository and navigate to `001-paste-markdown`.
2. Ensure you're on a macOS system.
3. Install dependencies:
   ```bash
   make setup
   ```

## Development Cycle
1. Run the build command:
   ```bash
   make build
   ```
2. Test the core conversion logic:
   ```bash
   make test
   ```

## Usage
1. Copy a rich-text snippet (e.g., from a web browser).
2. Run the application:
   ```bash
   ./bin/md-paste
   ```
3. Paste the Markdown in your editor.

## Common Scenarios
- **Direct Terminal Output**:
  ```bash
  ./bin/md-paste --stdout
  ```
- **View Help**:
  ```bash
  ./bin/md-paste --help
  ```
