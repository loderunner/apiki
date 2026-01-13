# apiki

**apiki** is a terminal-based environment variable manager that helps you organize, select, and apply environment variables across different projects and contexts.

## Features

- **Terminal User Interface (TUI)** - Navigate and manage variables with keyboard shortcuts
- **Fuzzy Search** - Quickly find variables by name or description
- **Radio Button Groups** - Multiple values for the same variable name (e.g., different database URLs)
- **`.env` File Integration** - Automatically discovers and loads `.env` files from your project directory
- **Environment Import** - Import variables from your current shell environment
- **Shell Integration** - Works seamlessly with bash, zsh, and fish

<!-- TODO: Add asciinema video (~20s) showing:
     1. Launch apiki
     2. Create SECRET_PASSWORD = hunter2
     3. Show DATABASE_URL radio group with dev/staging/prod alternatives
     4. Toggle between them with Space to demo radio behavior
     5. Press Enter to apply
-->
*Demo video coming soon*

## Installation

### Quick Install

```shell
curl -fsSL https://github.com/loderunner/apiki/releases/latest/download/install.sh | sh
```

After installation, close and reopen your terminal, or run:

```shell
source ~/.bashrc  # or ~/.zshrc, etc.
```

### Manual Installation

1. Download the latest release from [GitHub Releases](https://github.com/loderunner/apiki/releases)
2. Extract the archive for your platform
3. Move the binary to a directory in your PATH
4. Set up shell integration (see [Documentation](https://loderunner.github.io/apiki/docs/reference/shell-integration/))

## Quick Start

```shell
# Launch apiki
apiki

# Create entries, select variables, then press Enter to apply
# The shell commands are automatically evaluated
```

## Usage Example

```shell
# Launch apiki
apiki

# In the interface:
# 1. Press '+' to create a new entry
# 2. Enter: DATABASE_URL = postgres://localhost/mydb
# 3. Press Space to select it
# 4. Press Enter to apply

# Variables are now set in your shell
echo $DATABASE_URL
# Output: postgres://localhost/mydb
```

## Documentation

For complete documentation, visit: **[https://loderunner.github.io/apiki](https://loderunner.github.io/apiki)**

## License

Licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for details.
