---
title: 'Configuration'
weight: 2
---

apiki can be customized through command-line flags and environment variables.

## Variables File Location

Your saved variables are stored in a file. By default, this is `~/.apiki/variables.json`.

You can change the location:

**With a flag:**

```shell
apiki --variables-file /path/to/my-variables.json
apiki -f /path/to/my-variables.json
```

**With an environment variable:**

```shell
export APIKI_FILE=/path/to/my-variables.json
apiki
```

The flag takes precedence over the environment variable.

## Installation Directory

The installation directory contains the apiki binary and shell init scripts. By default:

- Linux/macOS: `~/.local/share/apiki`
- If `XDG_DATA_HOME` is set: `$XDG_DATA_HOME/apiki`

To install to a custom location, set `APIKI_DIR` before running the install script:

```shell
export APIKI_DIR="$HOME/custom/path"
curl -fsSL https://raw.githubusercontent.com/loderunner/apiki/main/scripts/install.sh | bash
```

## Command-Line Options

| Option                   | Description            |
| ------------------------ | ---------------------- |
| `--variables-file`, `-f` | Path to variables file |

## Environment Variables

| Variable     | Description            | Default                   |
| ------------ | ---------------------- | ------------------------- |
| `APIKI_FILE` | Path to variables file | `~/.apiki/variables.json` |
| `APIKI_DIR`  | Installation directory | `~/.local/share/apiki`    |

## Multiple Configurations

You can maintain separate variable files for different contexts:

```shell
# Work projects
apiki -f ~/.apiki/work.json

# Personal projects
apiki -f ~/.apiki/personal.json
```

Or set up shell aliases:

```shell
alias apiki-work='apiki -f ~/.apiki/work.json'
alias apiki-personal='apiki -f ~/.apiki/personal.json'
```

## Troubleshooting

### Shell Integration Not Working

1. Check that `APIKI_DIR` is set: `echo $APIKI_DIR`
2. Verify the init script exists: `ls "$APIKI_DIR/init.bash"` (or `.zsh`, `.fish`)
3. Ensure your shell config sources it
4. Restart your terminal or run `source ~/.bashrc` (or equivalent)

### Permission Errors

Make sure you have read/write access to:

- The directory containing your variables file
- The variables file itself

### Command Not Found

If `apiki` isn't recognized:

1. Check that the binary is in your PATH, or
2. Source your shell configuration: `source ~/.bashrc`
