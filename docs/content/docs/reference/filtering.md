---
title: "Filtering and Search"
weight: 23
---

# Filtering and Search

apiki includes a powerful fuzzy search feature to help you quickly find entries in large lists.

## Activating Filter Mode

Press `/` to open the filter input. The cursor moves to the filter bar at the bottom of the screen.

## How Filtering Works

apiki uses [fuzzy matching](https://github.com/sahilm/fuzzy) to find entries. This means you don't need to type exact matches - partial matches work too.

The search matches against:
- Entry names
- Entry labels
- For `.env` entries: the directory and filename (e.g., "project/.env")

## Example Searches

- Type `db` to find entries with "database" in the name or label
- Type `prod` to find production-related entries
- Type `api` to find API-related variables
- Type `project` to find entries from `.env` files in a project directory

## Match Highlighting

Matching characters are highlighted in **bright yellow** to show what matched your search query.

## Filter Actions

- **Type** - Search as you type (real-time filtering)
- **Enter** - Apply the filter and exit filter mode
- **Esc** - Clear the filter and return to normal view

## Filter Status

The filter bar shows:
- Current filter query
- Match count: `(X/Y entries)` where X is the number of matches and Y is the total

## Clearing Filters

You can clear the filter in several ways:

1. Press `Esc` while in filter mode
2. Press `Esc` while viewing filtered results (if filter is active)
3. Delete all text in the filter input and press `Enter`

## Cursor Persistence

When filtering, apiki tries to keep your cursor on the same entry if it's still visible after filtering. If the entry is filtered out, the cursor moves to the nearest visible entry above it.

{{< callout type="info" >}}
**Screenshot Coming Soon**

A screenshot showing the filter in action with highlighted matches will be added here.
{{< /callout >}}
