---
title: "Shell Integration"
weight: 27
---

# Shell Integration

apiki integrates with your shell to provide a seamless workflow for managing environment variables. This section explains how it works and how to use it effectively.

## How It Works

apiki outputs shell commands (`export` and `unset` statements) to stdout. These commands are designed to be evaluated by your shell to set the selected variables.

## The apiki Function

The init scripts provide an `apiki` shell function that wraps the binary:

{{< tabs items="Bash/Zsh,Fish" >}}

{{< tab >}}
```shell
apiki() {
  eval "$("${APIKI_DIR:-$HOME/.local/share/apiki}/apiki" "$@")"
}
```
{{< /tab >}}

{{< tab >}}
```fish
function apiki
  eval ("$APIKI_DIR/apiki" $argv)
end
```
{{< /tab >}}

{{< /tabs >}}

This function:
1. Runs the apiki binary
2. Captures the output (shell commands)
3. Evaluates the commands in the current shell

## Basic Usage

After shell integration is set up:

```shell
apiki
```

This:
1. Opens the TUI
2. Lets you select variables
3. Outputs commands when you quit
4. Evaluates the commands automatically

## Manual Evaluation

If you prefer to review the commands before applying them:

```shell
apiki > /tmp/apiki-commands.sh
cat /tmp/apiki-commands.sh
# Review the commands
source /tmp/apiki-commands.sh
```

Or pipe directly:

```shell
apiki | source /dev/stdin
```

## Output Format

apiki outputs standard shell commands:

```shell
export DATABASE_URL='postgres://localhost/mydb'
export API_KEY='secret123'
unset OLD_VAR
```

### Export Commands

Selected entries produce `export` commands:

```shell
export VAR_NAME='value'
```

Values are properly escaped for shell safety. Single quotes are escaped as `'\''`.

### Unset Commands

If a variable was set in your environment but no entry is selected, apiki outputs:

```shell
unset VAR_NAME
```

This ensures variables are cleared when deselected.

## Change Detection

apiki only outputs commands for variables that have **changed**:

- If a variable wasn't set and you select it → `export` command
- If a variable was set to value A and you select value B → `export` command with new value
- If a variable was set and you deselect it → `unset` command
- If a variable was set to value A and you keep value A → **no command** (no change)

This minimizes unnecessary commands and keeps your environment clean.

## Radio Button Groups

For entries with the same name (radio button groups), only the selected entry produces an export command. The others produce no output.

## Environment Snapshot

When apiki starts, it captures a snapshot of your current environment for all entry names. This snapshot is used to:

1. Determine which variables were originally set
2. Detect changes when you quit
3. Sync selection state (if an entry matches the environment, it's pre-selected)

## Shell Compatibility

apiki works with:

- **Bash** (3.0+)
- **Zsh** (all versions)
- **Fish** (all versions)

The output format is compatible with POSIX shells, though the init scripts are shell-specific.

## Troubleshooting

### Commands Not Applied

If variables aren't being set:

1. Check that shell integration is set up (see [Configuration](/docs/reference/configuration/))
2. Verify the `apiki` function exists: `type apiki`
3. Check for errors: `apiki 2>&1`

### Escaping Issues

If you have values with special characters:

- Single quotes are automatically escaped
- The shell function handles evaluation safely
- For complex values, review the output before evaluating

### Function Not Found

If you see "command not found":

1. Source your shell configuration: `source ~/.bashrc` (or equivalent)
2. Check that `APIKI_DIR` is set: `echo $APIKI_DIR`
3. Verify the init script exists: `ls "$APIKI_DIR/init.bash"`

## Advanced Usage

### Conditional Application

You can conditionally apply commands:

```shell
if apiki; then
  echo "Variables applied successfully"
else
  echo "apiki was cancelled"
fi
```

### Script Integration

In scripts, you might want to set variables without the TUI:

```shell
# Set variables directly (if apiki supported non-interactive mode)
# This is not currently supported, but you can work around it:
export DATABASE_URL='postgres://localhost/mydb'
```

For scripts, consider using `.env` files or setting variables directly.

## Best Practices

1. **Use descriptive labels** - Makes it easier to identify variables in the list
2. **Group related variables** - Use consistent naming patterns
3. **Review before applying** - Check the output if unsure
4. **Keep backups** - Your variables file is JSON, easy to backup
5. **Use .env files for projects** - Keep project-specific vars in `.env` files
