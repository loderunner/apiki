---
title: "Quick Start"
weight: 12
---

## Launching apiki

Simply run:

```shell
apiki
```

This opens the interactive terminal interface. You'll see a list of environment variables (initially empty if this is your first time).

## Creating Your First Variable

1. Press `+` to create a new variable
2. Fill in the form:
   - **Name**: The environment variable name (e.g., `DATABASE_URL`)
   - **Value**: The value to set (e.g., `postgres://localhost/mydb`)
   - **Label**: An optional description (e.g., `Local development database`)
3. Press `Enter` to save

The variable appears in your list and is saved for future sessions.

<!-- TODO: Add short asciinema video (~10s) showing:
     1. Press +
     2. Fill in Name, Value, Label fields
     3. Tab between fields
     4. Press Enter to save
     5. Variable appears in list
-->
*Video coming soon*

## Selecting Variables

- Use `↑`/`↓` or `j`/`k` to navigate the list
- Press `Space` to toggle selection

> [!NOTE]
> **Alternatives**: If you have multiple values for the same variable name, selecting one automatically deselects the others. This makes it easy to switch between dev, staging, and production configurations.

## Applying Changes

When you're done selecting:

1. Press `Enter` to apply and quit
2. apiki sets the selected variables in your shell:
   ```shell
   export DATABASE_URL='postgres://localhost/mydb'
   export API_KEY='secret123'
   ```

The variables are now set in your current shell session.

## Filtering

Press `/` to open the filter input. Type to search through variables by name or label. Matching characters are highlighted.

- Press `Enter` to apply the filter
- Press `Esc` to clear the filter

## Importing from Your Environment

You can import variables from your current shell:

1. Press `i` to enter import mode
2. Browse the list of current environment variables
3. Select the ones you want to save
4. Press `Enter` to confirm

Imported variables are added to your collection with the label "imported from environment".

## Working with .env Files

apiki automatically discovers `.env` files in your current directory and parent directories. Variables from these files appear in your list with a label showing their source.

You can:
- **Select** them to use their values
- **Save permanently** (press `=` on a .env variable) to add them to your collection

