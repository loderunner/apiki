---
title: "apiki"
weight: 1
---

**apiki** is a terminal-based environment variable manager that helps you organize, select, and apply environment variables across different projects and contexts. Instead of juggling multiple `.env` files or manually exporting variables, apiki lets you:

- **Organize** environment variables in a central location
- **Select** which variables to activate using a visual interface
- **Import** variables from existing `.env` files or your current shell environment
- **Apply** changes directly to your shell session

## Key Features

- **Visual Interface** – Navigate and manage variables with keyboard shortcuts
- **Fuzzy Search** – Quickly find variables by name or description
- **Alternatives** – Multiple values for the same variable (e.g., different database URLs for dev/staging/prod)
- **`.env` Integration** – Automatically discovers and loads `.env` files from your project
- **Environment Import** – Capture variables from your current shell
- **Shell Integration** – Works seamlessly with bash, zsh, and fish
- **Encryption** – Protect sensitive values at rest with password or keychain-based encryption

## Quick Start

```shell
# Install apiki
curl -fsSL https://raw.githubusercontent.com/loderunner/apiki/main/scripts/install.sh | bash

# Use apiki
apiki
```

For detailed installation instructions and usage, see the [Getting Started](/docs/getting-started/) section.
