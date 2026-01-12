---
title: "Browsing"
weight: 1
---

When you launch apiki, you see your environment variables in a scrollable list.

## The List View

Each row in the list shows:

- A checkbox indicating whether the variable is selected
- The variable name in bold
- An optional label in gray (e.g., "Local development database")

Variables with the same name are grouped together with visual connectors (`┌`, `├`, `└`), making it easy to see your alternatives at a glance.

## Navigating

Move through the list with:

- `↑` / `↓` arrow keys
- `j` / `k` (vim-style)

The current row is highlighted and marked with `>`.

## Filtering

Press `/` to start filtering. A search bar appears at the bottom of the screen.

Type to search—apiki uses fuzzy matching, so you don't need exact text. For example:

- Type `db` to find variables with "database" in the name or label
- Type `prod` to find production-related variables
- Type `api` to find API keys and URLs

### Exiting Filter Mode

- Press `Enter` to keep the filter active and return to navigating
- Press `Esc` to clear the filter and show all variables again
