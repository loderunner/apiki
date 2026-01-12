---
title: "Configuration"
weight: 26
---

# Configuration

apiki can be configured through command-line flags, environment variables, and file locations.

## Variables File Location

The variables file stores your apiki entries. The location is determined by priority:

1. **`--variables-file` flag** (highest priority)
2. **`APIKI_VARIABLES_FILE` environment variable**
3. **Default path**: `~/.apiki/variables.json`

### Using a Flag

```shell
apiki --variables-file /path/to/custom/variables.json
```

### Using Environment Variable

```shell
export APIKI_VARIABLES_FILE=/path/to/custom/variables.json
apiki
```

### Default Location

If neither flag nor environment variable is set, apiki uses:

```
~/.apiki/variables.json
```

The directory is created automatically if it doesn't exist.

## Installation Directory

The installation directory (`APIKI_DIR`) determines where apiki binaries and init scripts are located.

### Default Location

- **Linux/macOS**: `~/.local/share/apiki`
- **With XDG_DATA_HOME**: `$XDG_DATA_HOME/apiki`

### Custom Location

Set `APIKI_DIR` before running the install script:

```shell
export APIKI_DIR="$HOME/custom/path"
curl -fsSL https://raw.githubusercontent.com/loderunner/apiki/main/scripts/install.sh | bash
```

Or set it in your shell configuration:

```shell
export APIKI_DIR="$HOME/custom/path"
[ -s "$APIKI_DIR/init.bash" ] && . "$APIKI_DIR/init.bash"
```

## Shell Integration

apiki requires shell integration to work properly. The init scripts are located in `$APIKI_DIR/init.{bash,zsh,fish}`.

### Bash

Add to `~/.bashrc`:

```shell
export APIKI_DIR="$HOME/.local/share/apiki"
[ -s "$APIKI_DIR/init.bash" ] && . "$APIKI_DIR/init.bash"
```

### Zsh

Add to `~/.zshrc`:

```shell
export APIKI_DIR="$HOME/.local/share/apiki"
[ -s "$APIKI_DIR/init.zsh" ] && . "$APIKI_DIR/init.zsh"
```

### Fish

Add to `~/.config/fish/config.fish`:

```fish
set -gx APIKI_DIR "$HOME/.local/share/apiki"
source "$APIKI_DIR/init.fish"
```

## Command-Line Options

### `--variables-file` / `-f`

Specify a custom path to the variables file:

```shell
apiki --variables-file /path/to/variables.json
apiki -f /path/to/variables.json
```

### `version`

Print the version number:

```shell
apiki version
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `APIKI_VARIABLES_FILE` | Path to variables JSON file | `~/.apiki/variables.json` |
| `APIKI_DIR` | Installation directory | `~/.local/share/apiki` |

## File Formats

### Variables File

The variables file is JSON:

```json
[
  {
    "name": "DATABASE_URL",
    "value": "postgres://localhost/mydb",
    "label": "Local development database"
  }
]
```

The file is automatically created with proper formatting when you save entries.

### .env Files

apiki discovers `.env` files using standard format:

```
DATABASE_URL=postgres://localhost/mydb
API_KEY=secret123
```

See the [.env Integration](/docs/reference/dotenv/) section for details.

## Multiple Configurations

You can maintain multiple variable files for different contexts:

```shell
# Development
apiki --variables-file ~/.apiki/dev.json

# Production
apiki --variables-file ~/.apiki/prod.json
```

Or use environment variables:

```shell
export APIKI_VARIABLES_FILE=~/.apiki/dev.json
apiki
```

## Troubleshooting

### Variables File Not Found

If apiki can't find your variables file, it creates an empty one automatically. This is normal for first-time use.

### Permission Errors

Make sure you have read/write permissions to:
- The variables file directory
- The variables file itself

### Shell Integration Not Working

1. Verify `APIKI_DIR` is set correctly
2. Check that the init script exists: `ls "$APIKI_DIR/init.bash"`
3. Ensure your shell configuration sources the init script
4. Restart your terminal or run `source ~/.bashrc` (or equivalent)
