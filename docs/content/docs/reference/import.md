---
title: "Importing from Environment"
weight: 25
---

# Importing from Environment

apiki can import environment variables from your current shell environment, making it easy to capture existing variables and add them to your apiki file.

{{< callout type="info" >}}
**Screenshot Coming Soon**

A screenshot showing the import mode interface will be added here.
{{< /callout >}}

## Entering Import Mode

Press `i` from the main list view to enter import mode. The interface changes to show all environment variables from your current shell.

## Import Mode Behavior

In import mode:

- **All current environment variables** are displayed
- You can **select multiple entries** to import
- **Radio button behavior is disabled** - you can select multiple entries with the same name
- The title changes to "Import from Environment"

## Selecting Variables to Import

1. Navigate through the list using `↑`/`↓` or `j`/`k`
2. Press `Space` to toggle selection
3. Select all variables you want to import

You can select as many variables as you want, including multiple entries with the same name.

## Confirming Import

1. Press `Enter` to confirm the import
2. A confirmation dialog shows how many entries will be imported
3. Press `y` or `Enter` to proceed, or `n`/`Esc` to cancel

## Imported Entry Properties

Imported entries are created with:

- **Name** - The original environment variable name
- **Value** - The current value from your environment
- **Label** - Set to "imported from environment"
- **Selected** - Set to `true` (so they're immediately active)

After import, entries are:
- Added to your apiki variables file
- Sorted alphabetically with other entries
- Available for future use

## Canceling Import

Press `Esc` to cancel import mode and return to the normal list view. No changes are made.

## Use Cases

### Capturing Current Environment

If you have variables set in your shell that you want to save:

```shell
export DATABASE_URL="postgres://localhost/mydb"
export API_KEY="secret123"
apiki  # Press 'i' to import these
```

### Migrating from Manual Setup

If you've been manually exporting variables, import them all at once:

1. Run `apiki`
2. Press `i` to enter import mode
3. Select all relevant variables
4. Confirm import

### Creating a Snapshot

Import can be useful for creating a snapshot of your current environment state for documentation or backup purposes.

## Limitations

- Imported entries use the **current value** at import time
- Values are **static** - they won't update if the environment changes
- Very large environments (100+ variables) may be slow to import

{{< callout >}}
**Tip**: After importing, review the entries and add descriptive labels to make them easier to identify later.
{{< /callout >}}
