---
title: "Selecting"
weight: 2
---

Selecting determines which variables will be set in your shell when you apply changes.

## Toggling Selection

Press `Space` to toggle the current variable on or off.

- `⦿` means the variable is selected and will be exported
- `◯` means the variable is not selected

## Alternatives

When you have multiple values for the same variable name (e.g., different `DATABASE_URL` values for dev, staging, and production), they form a radio-button group.

**Only one alternative can be selected at a time.** Selecting one automatically deselects the others in the group. This ensures you don't accidentally export conflicting values.

For example:

```
┌ ◯ DATABASE_URL  postgres://localhost/dev       (dev)
├ ⦿ DATABASE_URL  postgres://staging.example     (staging)
└ ◯ DATABASE_URL  postgres://prod.example        (production)
```

Here, the staging value is selected. Selecting the production row would automatically deselect staging.

<!-- TODO: Add screenshot (svg-term still frame) showing:
     - DATABASE_URL radio group with 3 alternatives (dev/staging/prod)
     - Visual connectors (┌, ├, └) linking them
     - One selected (⦿), two deselected (◯)
     - Maybe another unrelated variable for context
     Workflow: asciinema rec → convert to v2 → svg-term --at <ms>
-->
*Screenshot coming soon*

## Applying Changes

When you're done selecting:

- Press `Enter` to apply your changes and exit
- Press `q` or `Ctrl+C` to quit without applying anything

When you apply, apiki sets the selected variables in your current shell session. Variables you deselected are unset.

## What Gets Applied

apiki only changes variables that are different from your current environment:

- If you select a variable that wasn't set → it gets exported
- If you select a different value for an existing variable → it gets updated
- If you deselect a variable that was set → it gets unset
- If nothing changed → no commands are run

This keeps your shell clean and avoids unnecessary churn.
