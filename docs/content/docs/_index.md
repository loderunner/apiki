---
title: "apiki"
weight: 1
---

# apiki

**apiki** is a terminal-based environment variable manager that helps you organize, select, and apply environment variables across different projects and contexts.

## What is apiki?

apiki provides a clean, interactive terminal interface for managing environment variables. Instead of juggling multiple `.env` files or manually exporting variables, apiki lets you:

- **Organize** environment variables in a central location
- **Select** which variables to activate using a visual interface
- **Import** variables from existing `.env` files or your current shell environment
- **Apply** changes by outputting shell commands that you can evaluate

## Key Features

- **Terminal User Interface (TUI)** - Navigate and manage variables with keyboard shortcuts
- **Fuzzy Search** - Quickly find variables by name or description
- **Radio Button Groups** - Multiple values for the same variable name (e.g., different database URLs)
- **`.env` File Integration** - Automatically discovers and loads `.env` files from your project directory
- **Environment Import** - Import variables from your current shell environment
- **Shell Integration** - Works seamlessly with bash, zsh, and fish

## Quick Start

```shell
# Install apiki
curl -fsSL https://raw.githubusercontent.com/loderunner/apiki/main/scripts/install.sh | bash

# Use apiki
apiki
```

For detailed installation instructions and usage, see the [Getting Started](/docs/getting-started/) section.
