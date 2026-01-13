---
title: "Importing"
weight: 5
---

apiki can capture variables from your current shell environment and save them to your collection.

## Entering Import Mode

Press `i` from the main list to enter import mode. The interface switches to show all environment variables currently set in your shell.

## Selecting Variables to Import

1. Navigate the list with `↑`/`↓` or `j`/`k`
2. Press `Space` to select variables you want to import
3. Select as many as you need

## Confirming the Import

1. Press `Enter` when you've selected everything you want
2. Confirm when prompted
3. The variables are added to your collection

<!-- TODO: Add asciinema video (~12s) showing import mode:
     1. Press i to enter import mode
     2. Show list of current shell environment variables
     3. Navigate and select a few with Space
     4. Press Enter to confirm
     5. Back to main list, imported vars appear with "imported from environment" label
     NOTE: Record in a clean/dummy environment (Docker or env -i) to avoid leaking personal info
-->
*Video coming soon*

Imported variables are automatically selected and labeled "imported from environment" so you can identify them later. You can edit them afterwards to add more descriptive labels.

## Canceling

Press `Esc` to cancel import mode and return to the main list without importing anything.

## When to Use Import

**Migrating from manual exports:**
If you've been setting variables with `export VAR=value` in your shell config, import them to apiki so you can manage them visually.

**Capturing a working configuration:**
When you have a shell session with the right variables set, import them to save that configuration for later.

**Bootstrapping a new setup:**
Import variables from a working machine to quickly set up a new environment.
