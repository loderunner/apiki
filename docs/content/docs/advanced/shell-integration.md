---
title: "Shell Integration"
weight: 3
---

apiki integrates with your shell to set environment variables in your current session.

## How It Works

When you quit apiki (with `Enter`), it outputs shell commands to stdout:

```shell
export DATABASE_URL='postgres://localhost/mydb'
export API_KEY='secret123'
unset OLD_VAR
```

The shell integration wraps apiki so these commands are automatically evaluated in your current shell.

## Supported Shells

- **Bash** (3.0+)
- **Zsh** (all versions)
- **Fish** (all versions)

## Output Format

**Selected variables** produce `export` commands:

```shell
export VAR_NAME='value'
```

Values are properly escaped for shell safety.

**Deselected variables** that were previously set produce `unset` commands:

```shell
unset VAR_NAME
```

## Manual Evaluation

If you want to review commands before applying, or integrate apiki into a script:

```shell
# Preview the commands
"$APIKI_DIR/apiki" > /tmp/apiki-commands.sh
cat /tmp/apiki-commands.sh

# Apply after review
source /tmp/apiki-commands.sh
```

Or pipe directly:

```shell
eval "$("$APIKI_DIR/apiki")"
```

## Shell Setup

The init scripts set up a wrapper function and optionally enable auto-restore. For reference:

{{< tabs items="Bash/Zsh,Fish" >}}

{{< tab >}}

```shell
# Auto-restore apiki state on shell startup (opt-in via APIKI_AUTO_RESTORE)
# Only runs in the first shell, not subshells (APIKI_RESTORED marker)
if [ -n "$APIKI_AUTO_RESTORE" ] && [ -z "$APIKI_RESTORED" ]; then
  eval "$("${APIKI_DIR:-$HOME/.local/share/apiki}/apiki" restore 2>/dev/null)"
  export APIKI_RESTORED=1
fi

apiki() {
  eval "$("${APIKI_DIR:-$HOME/.local/share/apiki}/apiki" "$@")"
}
```

{{< /tab >}}

{{< tab >}}

```fish
set -q APIKI_DIR; or set APIKI_DIR "$HOME/.local/share/apiki"

# Auto-restore apiki state on shell startup (opt-in via APIKI_AUTO_RESTORE)
# Only runs in the first shell, not subshells (APIKI_RESTORED marker)
if set -q APIKI_AUTO_RESTORE; and not set -q APIKI_RESTORED
  eval ("$APIKI_DIR/apiki" restore 2>/dev/null)
  set -gx APIKI_RESTORED 1
end

function apiki
  eval ("$APIKI_DIR/apiki" $argv)
end
```

{{< /tab >}}

{{< /tabs >}}

## The `restore` Command

The `apiki restore` command restores the variables you had selected the last time you used apiki. This is useful when you open a new terminal and want to restore your environment.

```shell
apiki restore
```

This outputs export commands for your previously selected variables, which are then evaluated by the shell integration wrapper.

**Example:**

```shell
$ apiki restore
export DATABASE_URL='postgres://localhost/mydb'
export API_KEY='secret123'
```

The variables are now set in your current shell session.

## Auto-Restore

You can enable automatic variable restore when opening a new terminal. This runs `apiki restore` automatically on shell startup.

**To enable:**

Set `APIKI_AUTO_RESTORE=1` in your shell configuration file (before sourcing the apiki init script):

```shell
# In ~/.bashrc, ~/.zshrc, or ~/.config/fish/config.fish
export APIKI_AUTO_RESTORE=1
source "$APIKI_DIR/init.bash"  # or init.zsh, init.fish
```

**How it works:**

- When you open a new terminal, apiki automatically restores your variables
- Only runs once per terminal session (subshells inherit variables from the parent shell)
- For encrypted files: password or Touch ID prompt only appears in the first shell, not in subshells

**Note:** Auto-restore is opt-in. If you don't set `APIKI_AUTO_RESTORE`, you can still manually run `apiki restore` whenever you need it.

## Best Practices

- **Use descriptive labels** – Makes variables easier to find
- **Group related variables** – Use consistent naming (e.g., `DB_*`, `AWS_*`)
- **Use .env files for projects** – Keep project-specific config in `.env` files
