# dict-gen - Dictionary Generator CLI

A command-line tool for managing the Te Reo Bot dictionary using SQLite database.

## Overview

The `dict-gen` tool manages the word-of-the-day dictionary by:
- Storing all Māori words in a SQLite database (not just 365-366)
- Validating that exactly 366 words are assigned to days (1-366)
- Generating the `dictionary.json` file used by the HTTP server

## Installation

Build the CLI tool:

```bash
go build -o dict-gen ./cmd/dict-gen/
```

## Usage

### migrate - Import dictionary.json to SQLite

Import the existing dictionary.json file into the SQLite database:

```bash
dict-gen migrate --input=./cmd/server/dictionary.json --db=./data/words.db
```

Flags:
- `--input`: Path to input dictionary.json file (default: `./cmd/server/dictionary.json`)
- `--db`: Path to SQLite database file (default: `./data/words.db`)

### validate - Check database integrity

Validate that the database contains exactly 366 unique day indexes (1-366):

```bash
dict-gen validate --db=./data/words.db
```

Flags:
- `--db`: Path to SQLite database file (default: `./data/words.db`)

### generate - Generate dictionary.json

Generate dictionary.json from the SQLite database:

```bash
dict-gen generate --output=./cmd/server/dictionary.json --db=./data/words.db
```

Flags:
- `--output`: Path to output dictionary.json file (default: `./cmd/server/dictionary.json`)
- `--db`: Path to SQLite database file (default: `./data/words.db`)
- `--compact`: Generate compact JSON without indentation (default: false)

The generate command automatically validates the database before generating. If validation fails, no file is generated.

## Workflow

### Initial Setup

1. **Migrate existing dictionary**:
   ```bash
   dict-gen migrate
   ```
   This imports all 365-366 words from dictionary.json into SQLite.

2. **Validate the database**:
   ```bash
   dict-gen validate
   ```
   Ensure all 366 day indexes (1-366) have exactly one word assigned.

3. **Generate dictionary.json**:
   ```bash
   dict-gen generate
   ```
   Create the dictionary.json file that the HTTP server uses.

### Regular Workflow

When adding or editing words:

1. Edit the SQLite database directly (using DB Browser for SQLite or SQL commands)
2. Run `dict-gen validate` to check integrity
3. Run `dict-gen generate` to update dictionary.json
4. Commit both the database and generated JSON:
   ```bash
   git add data/words.db cmd/server/dictionary.json
   git commit -m "Update dictionary with new words"
   ```

## Database Schema

The SQLite database has a `words` table:

| Column | Type | Description |
|--------|------|-------------|
| id | INTEGER PRIMARY KEY | Auto-increment ID |
| day_index | INTEGER UNIQUE | Day of year (1-366), nullable |
| word | TEXT NOT NULL | Māori word |
| meaning | TEXT NOT NULL | English meaning |
| link | TEXT | Optional link |
| photo | TEXT | Optional photo filename |
| photo_attribution | TEXT | Optional photo attribution |
| created_at | DATETIME | Creation timestamp |
| updated_at | DATETIME | Last update timestamp |
| is_active | BOOLEAN | Soft delete flag |

Words with `day_index = NULL` are stored but not included in dictionary.json generation.

## Architecture

See `specs/sqlite_dictionary_architecture/design.md` for the complete architecture documentation.

## Benefits

1. **Store unlimited words**: Database can hold thousands of words, not just 366
2. **Better validation**: Ensures exactly 366 unique day indexes before generation
3. **Version control**: Track all words over time, not just current 366
4. **Easier editing**: Use DB tools instead of manually editing JSON
5. **Data integrity**: Database constraints prevent duplicate day indexes
6. **Atomic updates**: Generated JSON is written atomically (temp file + rename)

## Troubleshooting

### Validation fails with missing indexes

If validation shows missing day indexes:

```bash
❌ Validation failed!
   - Missing indexes: [100, 200, 366]
```

Solution: Assign words to the missing day indexes in the database.

### Migration reports fewer than 366 words

The current dictionary.json may have fewer than 366 entries. This is expected if day index 366 is missing (non-leap year). You'll need to add a word for day 366 to complete the set.

### Database file doesn't exist

The database file is created automatically on first `migrate` command. The database file itself is not committed to git (see `.gitignore`), but can be committed if desired for team collaboration.

## Testing

Run all tests:

```bash
go test ./pkg/...
```

Test specific package:

```bash
go test ./pkg/repository/... -v
go test ./pkg/migration/... -v
go test ./pkg/validator/... -v
go test ./pkg/generator/... -v
```

Total test coverage: 46 tests across all packages.
