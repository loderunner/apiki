---
title: "Creating and Managing Variables"
weight: 3
---

apiki stores your variables so you can reuse them across sessions.

## Creating a Variable

1. Press `+` to open the creation form
2. Fill in the fields:
   - **Name** – The environment variable name (e.g., `DATABASE_URL`)
   - **Value** – The value to set (e.g., `postgres://localhost/mydb`)
   - **Label** – An optional description to help you remember what this is for
3. Press `Enter` to save

The new variable appears in your list and is saved for future sessions.

### Form Navigation

- `Tab` or `↓` to move to the next field
- `Shift+Tab` or `↑` to move to the previous field
- `Enter` to save (when on the last field)
- `Esc` to cancel without saving

## Editing a Variable

1. Navigate to the variable you want to edit
2. Press `=` to open the edit form
3. Make your changes
4. Press `Enter` to save

{{< callout >}}
Variables from `.env` files can't be edited directly. To modify one, first save it permanently (see [.env Files](/docs/using-apiki/dotenv/)).
{{< /callout >}}

## Deleting a Variable

1. Navigate to the variable you want to delete
2. Press `-`, `Delete`, or `Backspace`
3. Confirm the deletion

{{< callout type="warning" >}}
Variables from `.env` files can't be deleted from apiki—they come from the files in your project. To remove them, edit the `.env` file directly.
{{< /callout >}}

## Creating Alternatives

To have multiple values for the same variable (e.g., different database URLs for different environments):

1. Create a new variable with `+`
2. Use the same **Name** as an existing variable
3. Give it a different **Value** and a descriptive **Label**

The variables will be grouped together, and only one can be selected at a time. See [Selecting](/docs/using-apiki/selecting/) for details on how alternatives work.
