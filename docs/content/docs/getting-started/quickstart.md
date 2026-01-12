---
title: "Quick Start"
weight: 12
---

# Quick Start

This guide will walk you through your first steps with apiki. You'll learn how to create entries, select variables, and apply them to your shell.

{{< callout type="info" >}}
**Video Tutorial Coming Soon**

A video walkthrough demonstrating apiki's features will be available here soon.
{{< /callout >}}

## Launching apiki

Simply run:

```shell
apiki
```

This opens the interactive terminal interface. You'll see a list of environment variables (initially empty if you haven't created any yet).

## Creating Your First Entry

1. Press `+` to create a new entry
2. Fill in the form:
   - **Name**: The environment variable name (e.g., `DATABASE_URL`)
   - **Value**: The value to set (e.g., `postgres://localhost/mydb`)
   - **Label**: An optional description (e.g., `Local development database`)
3. Press `Enter` to save

The entry is saved to `~/.apiki/variables.json` and appears in your list.

## Selecting Variables

- Use `↑`/`↓` or `j`/`k` to navigate the list
- Press `Space` to toggle selection of an entry
- Selected entries are marked with `⦿`, unselected with `◯`

{{< callout >}}
**Radio Button Behavior**: If you have multiple entries with the same name, selecting one automatically deselects the others. This is useful for managing different environments (dev, staging, prod) for the same variable.
{{< /callout >}}

## Applying Changes

When you're done selecting variables:

1. Press `Enter` (or `q` to quit without applying)
2. apiki outputs shell commands to stdout:
   ```shell
   export DATABASE_URL='postgres://localhost/mydb'
   export API_KEY='secret123'
   unset OLD_VAR
   ```
3. Evaluate these commands in your shell:
   ```shell
   eval "$(apiki)"
   ```

The variables are now set in your current shell session.

## Filtering Entries

Press `/` to open the filter input. Type to search through entries by name or label. Matching characters are highlighted.

- Press `Enter` to apply the filter
- Press `Esc` to clear the filter

## Importing from Environment

You can import variables from your current shell environment:

1. Press `i` to enter import mode
2. Browse the list of current environment variables
3. Select the ones you want to add to apiki
4. Press `Enter` to confirm import

Imported entries are saved to your apiki file with the label "imported from environment".

## Importing from .env Files

apiki automatically discovers `.env` files in your current directory and parent directories. Entries from these files appear in your list with a label like "from project/.env".

You can:
- **Select** them to use their values
- **Promote** them (press `=` on a `.env` entry) to add them permanently to your apiki file

## Next Steps

Now that you know the basics, explore the [Reference Documentation](/docs/reference/) to learn about all features in detail.
