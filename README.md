# sawmill

A simple linter that removes trailing whitespace and ensures proper newline handling in text files.

*Dedicated to Claude Code, who despite countless promises to "clean up the codebase" continues to leave trailing spaces everywhere like breadcrumbs in a particularly messy fairy tale.*

## Features

- Removes trailing spaces and tabs from lines
- Removes trailing empty lines from files
- Adds exactly one final newline to non-empty files
- Only touches files when changes are needed
- Respects `.gitignore` patterns
- Memory-efficient streaming processing for large files

## Usage

```bash
# Process default file types in current directory
./sawmill

# Process specific file patterns
./sawmill --pattern="*.go,*.js"

# Ignore gitignore and process all matching files
./sawmill -u

# Enable verbose output
./sawmill --verbose
```

## Build

```bash
go build -o sawmill
```