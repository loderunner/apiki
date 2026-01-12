---
title: ".env File Integration"
weight: 24
---

# .env File Integration

apiki automatically discovers and loads `.env` files from your current directory and parent directories, making it easy to work with project-specific environment variables.

## How It Works

When you run `apiki`, it:

1. Starts from your current working directory (`PWD`)
2. Searches upward through parent directories
3. Finds all files matching `.env` or `.env.*` (e.g., `.env.local`, `.env.production`)
4. Parses each file and adds entries to your list

Entries from `.env` files are marked with a label showing their source:

```
DATABASE_URL  from project/.env
API_KEY       from project/.env.local
```

The label format is: `from <dirname>/<filename>`.

## Entry Properties

`.env` entries have special properties:

- **Read-only** - You cannot edit or delete them directly
- **Source tracking** - The `SourceFile` field stores the full path
- **Automatic discovery** - They appear automatically when you run apiki in a directory with `.env` files

## Promoting .env Entries

To add a `.env` entry permanently to your apiki file:

1. Navigate to the `.env` entry
2. Press `=` to promote it
3. Confirm the promotion dialog
4. The entry opens in edit mode - modify if needed
5. Press `Enter` to save

After promotion, the entry becomes a regular apiki entry and is stored in `~/.apiki/variables.json`. It will no longer show the "from ..." label.

## Use Cases

### Project-Specific Variables

Keep project-specific variables in `.env` files:

```shell
# project/.env
DATABASE_URL=postgres://localhost/projectdb
REDIS_URL=redis://localhost:6379
```

These appear automatically when you run `apiki` from the project directory.

### Environment-Specific Files

Use `.env.*` files for different environments:

```shell
# .env.development
API_URL=http://localhost:3000

# .env.production
API_URL=https://api.example.com
```

All matching files are discovered and loaded.

### Multiple Projects

When working with multiple projects, apiki shows entries from all discovered `.env` files. You can:

- Select variables from different projects
- Promote frequently-used ones to your main apiki file
- Use filtering to find project-specific entries

## File Discovery Order

Files are discovered from deepest to shallowest directory:

```
/home/user/projects/myapp/.env          (discovered first)
/home/user/projects/.env                (discovered second)
/home/user/.env                         (discovered third)
/home/.env                              (discovered last)
```

If multiple files define the same variable, entries from deeper directories appear first in the list.

## File Format

apiki uses the standard `.env` file format:

```
# Comments are supported
DATABASE_URL=postgres://localhost/mydb
API_KEY=secret123

# Empty lines are ignored
MULTILINE_VAR="value with
multiple lines"
```

The format is parsed using [godotenv](https://github.com/joho/godotenv), which follows the same rules as most `.env` parsers.

## Limitations

- `.env` entries cannot be edited or deleted directly
- Changes to `.env` files require restarting apiki to see updates
- Very large `.env` files may slow down the interface

{{< callout >}}
**Tip**: If you frequently use variables from a `.env` file, promote them to your apiki file for easier management.
{{< /callout >}}
