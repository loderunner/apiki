---
title: "Managing Entries"
weight: 22
---

# Managing Entries

Entries are the core unit of apiki. Each entry represents an environment variable with a name, value, and optional label.

## Entry Structure

An entry consists of:

- **Name** - The environment variable name (e.g., `DATABASE_URL`)
- **Value** - The value to set when selected (e.g., `postgres://localhost/mydb`)
- **Label** - Optional human-readable description (e.g., `Local development database`)
- **Selected** - Whether this entry is currently selected (not stored in file)

## Creating Entries

### From the Interface

1. Press `+` to create a new entry
2. Fill in the form fields:
   - Name (required)
   - Value (required)
   - Label (optional)
3. Press `Enter` to save

The entry is saved to your apiki variables file (`~/.apiki/variables.json` by default).

### From .env Files

Entries from `.env` files appear automatically in your list. To add them permanently:

1. Navigate to the `.env` entry
2. Press `=` to promote it
3. Confirm the promotion
4. Edit the entry if needed (it opens in edit mode)

### By Importing

See the [Import](/docs/reference/import/) section for details on importing from your shell environment.

## Editing Entries

1. Navigate to the entry you want to edit
2. Press `=` to edit
3. Modify the fields as needed
4. Press `Enter` to save

{{< callout >}}
**Note**: You can only edit apiki entries (those stored in your variables file). Entries from `.env` files must be promoted first.
{{< /callout >}}

## Deleting Entries

1. Navigate to the entry you want to delete
2. Press `-`, `Delete`, or `Backspace`
3. Confirm the deletion

{{< callout type="warning" >}}
**Warning**: You cannot delete entries from `.env` files. Only entries stored in your apiki variables file can be deleted.
{{< /callout >}}

## Radio Button Groups

When multiple entries share the same variable name, they form a **radio button group**. Only one entry in a group can be selected at a time.

This is useful for managing different environments:

```
┌ DATABASE_URL  postgres://localhost/dev     (dev)
├ DATABASE_URL  postgres://staging.example   (staging)
└ DATABASE_URL  postgres://prod.example      (production)
```

Selecting one automatically deselects the others. This ensures only one value is exported for each variable name.

## Entry Storage

apiki entries are stored in JSON format at `~/.apiki/variables.json`:

```json
[
  {
    "name": "DATABASE_URL",
    "value": "postgres://localhost/mydb",
    "label": "Local development database"
  },
  {
    "name": "API_KEY",
    "value": "secret123",
    "label": "Development API key"
  }
]
```

The file is automatically created when you save your first entry.

## Sorting

Entries are automatically sorted alphabetically by name (case-insensitive), then by label. This ensures a consistent, predictable order in the interface.

{{< callout type="info" >}}
**Screenshots Coming Soon**

Screenshots showing the entry creation workflow, editing, and radio button groups will be added here.
{{< /callout >}}
