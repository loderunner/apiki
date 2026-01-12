---
title: "User Interface"
weight: 21
---

# User Interface

apiki provides a terminal-based user interface (TUI) built with [Bubble Tea](https://github.com/charmbracelet/bubbletea). This section covers navigation and keyboard shortcuts.

{{< callout type="info" >}}
**Screenshots Coming Soon**

Screenshots of the interface will be added here to illustrate the different views and modes.
{{< /callout >}}

## Main List View

The main view displays all your environment variable entries in a scrollable list. Each entry shows:

- **Checkbox** (`⦿` selected, `◯` unselected) - Whether the variable will be exported
- **Name** - The environment variable name (bold)
- **Label** - Optional description (gray, italic)
- **Group indicators** - Visual connectors for entries with the same name

Entries with the same variable name are visually grouped with tree-like connectors (`┌`, `├`, `└`).

## Keyboard Shortcuts

### Navigation

| Key | Action |
|-----|--------|
| `↑` / `k` | Move cursor up |
| `↓` / `j` | Move cursor down |
| `Space` | Toggle selection of current entry |
| `/` | Open filter/search |

### Actions

| Key | Action |
|-----|--------|
| `+` | Create new entry |
| `=` | Edit entry (or promote `.env` entry) |
| `-` / `Delete` / `Backspace` | Delete entry (apiki entries only) |
| `i` | Import from environment |
| `Enter` | Apply changes and quit |
| `q` / `Ctrl+C` | Cancel and quit without applying |
| `Esc` | Clear filter or cancel current action |

### Filter Mode

When filtering (after pressing `/`):

| Key | Action |
|-----|--------|
| Type | Search entries (fuzzy matching) |
| `Enter` | Apply filter |
| `Esc` | Clear filter and exit filter mode |

## Form View

When creating or editing an entry, you'll see a form with three fields:

1. **Name** - Environment variable name (required)
2. **Value** - Variable value (required)
3. **Label** - Optional description

### Form Navigation

| Key | Action |
|-----|--------|
| `Tab` / `↓` | Move to next field |
| `Shift+Tab` / `↑` | Move to previous field |
| `Enter` | Save entry (when on last field) |
| `Esc` | Cancel and return to list |

## Confirmation Dialogs

When deleting an entry or promoting a `.env` entry, apiki shows a confirmation dialog:

| Key | Action |
|-----|--------|
| `y` / `Enter` | Confirm action |
| `n` / `Esc` | Cancel action |

## Visual Indicators

- **`▲`** - More entries above the visible area
- **`▼`** - More entries below the visible area
- **`>`** - Current cursor position
- **Highlighted text** - Matches in filter results

## Color Scheme

apiki uses colors to distinguish different elements:

- **Bright blue** - Titles and headers
- **Bright green** - Selected entries
- **Gray** - Unselected entries, labels, metadata
- **Bright yellow** - Filter match highlights
- **Bright red** - Error messages
- **Bright cyan** - Filter bar
