---
title: ".env Files"
weight: 4
---

apiki automatically discovers `.env` files from your project, so you can select and apply variables from them without copying them manually.

## Automatic Discovery

When you run `apiki`, it searches for `.env` files starting from your current directory and moving up through parent directories. It finds:

- `.env`
- `.env.local`
- `.env.development`
- `.env.production`
- Any file matching `.env.*`

Variables from these files appear in your list with a label showing where they came from (e.g., "from myproject/.env").

## Using .env Variables

You can select and apply variables from `.env` files just like any other variable. Toggle selection with `Space` and apply with `Enter`.

This is useful when you want to:

- Quickly load project-specific configuration
- Switch between different `.env.*` files (e.g., `.env.development` vs `.env.production`)
- Combine variables from multiple sources

## Limitations

Variables from `.env` files are **read-only** in apiki:

- You can't edit them directly—edit the `.env` file instead
- You can't delete them—remove them from the `.env` file
- Changes to `.env` files require restarting apiki to see updates

## Saving .env Variables Permanently

If you want to keep a variable from a `.env` file in your apiki collection (so it's available even outside that project):

1. Navigate to the `.env` variable
2. Press `=` to save it permanently
3. Confirm when prompted
4. Edit the variable if needed (it opens in the edit form)
5. Press `Enter` to save

The variable is now part of your personal collection and will appear regardless of which directory you're in.

## Multiple Projects

When you work in different project directories, apiki shows the `.env` files relevant to each one. Your personal variables are always visible, but `.env` variables change based on where you run apiki.

This makes it easy to switch contexts—just `cd` to a project and run `apiki` to see that project's environment configuration.
