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

<!-- TODO: Add screenshot (svg-term still frame) showing the list view:
     - Several variables with checkboxes (⦿/◯)
     - A radio group with connectors (┌├└)
     - Variable names in bold, labels in gray
     - Cursor marker (>) on one row
     Workflow: asciinema rec → convert to v2 → svg-term --at <ms>
-->
*Screenshot coming soon*

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

<!-- TODO: Add asciinema video (~15s) showing real-time filtering:
     Setup: ~15-20 variables with labels like:
       - DATABASE_URL (labels: "local dev", "staging", "production")
       - STRIPE_API_KEY (labels: "test mode", "production")
       - AWS_ACCESS_KEY_ID (label: "prod account")
       - GITHUB_API_TOKEN, REDIS_URL, SECRET_PASSWORD (hunter2), etc.
     Flow:
       1. Press / to open filter
       2. Type "api" → narrows to API-related vars (filtering by name)
       3. Backspace to empty → full list returns
       4. Type "prod" → shows vars with "prod"/"production" in LABEL
     Shows: real-time filtering, backspace widening, filtering by label
-->
*Video coming soon*

### Exiting Filter Mode

- Press `Enter` to keep the filter active and return to navigating
- Press `Esc` to clear the filter and show all variables again
