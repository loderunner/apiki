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

The init scripts set up a wrapper function. For reference:

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

## Best Practices

- **Use descriptive labels** – Makes variables easier to find
- **Group related variables** – Use consistent naming (e.g., `DB_*`, `AWS_*`)
- **Use .env files for projects** – Keep project-specific config in `.env` files
