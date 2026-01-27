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
- `--dry-run`: Preview migration without modifying database (default: `false`)

#### Dry-run Mode

Preview migration without modifying the database:

```bash
dict-gen migrate --input=./cmd/server/dictionary.json --dry-run
```

This will:
- Parse the JSON file
- Show number of words to import
- Display sample words
- Check for duplicate or missing day indexes
- Exit without touching the database

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

## Backup & Recovery

### Automatic Backups

The `dict-gen` tool automatically creates backups before potentially destructive operations:

**Migration backups:**
- Created before running `dict-gen migrate`
- Format: `words.db.backup.YYYYMMDD-HHMMSS`
- Location: Same directory as database file
- Retention: Last 7 days kept automatically

**Generation backups:**
- Created before running `dict-gen generate`
- Format: `dictionary.json.backup.YYYYMMDD-HHMMSS`
- Location: Same directory as output file

### Rollback Procedures

#### Rollback Database Migration

If a migration fails or produces incorrect results:

1. **List available backups:**
   ```bash
   ls -lh ./data/words.db.backup.*
   ```

2. **Restore from backup:**
   ```bash
   cp ./data/words.db.backup.YYYYMMDD-HHMMSS ./data/words.db
   ```

3. **Verify restoration:**
   ```bash
   dict-gen validate
   ```

4. **Re-generate dictionary.json from restored database:**
   ```bash
   dict-gen generate
   ```

#### Rollback dictionary.json Generation

If generated dictionary.json is corrupted or incorrect:

1. **List available backups:**
   ```bash
   ls -lh ./cmd/server/dictionary.json.backup.*
   ```

2. **Restore from backup:**
   ```bash
   cp ./cmd/server/dictionary.json.backup.YYYYMMDD-HHMMSS ./cmd/server/dictionary.json
   ```

3. **Verify JSON is valid:**
   ```bash
   python -m json.tool ./cmd/server/dictionary.json > /dev/null && echo "Valid JSON"
   ```

#### Complete System Rollback

To revert to previous working state:

1. Restore database from backup (see above)
2. Restore dictionary.json from backup (see above)
3. Restart HTTP server if running

#### Emergency: Revert to JSON-Only Workflow

If SQLite migration causes issues, you can temporarily revert to the old JSON-editing workflow:

1. Keep the last working `dictionary.json`
2. Stop using `dict-gen` commands
3. Edit `dictionary.json` directly
4. Use existing git hooks for validation
5. Investigate migration issues separately

### Backup Best Practices

1. **Manual backups before major changes:**
   ```bash
   cp ./data/words.db ./data/words.db.manual-backup-$(date +%Y%m%d)
   ```

2. **Git commit generated files:**
   After successful migration/generation:
   ```bash
   git add ./cmd/server/dictionary.json
   git commit -m "Update dictionary from SQLite"
   ```

3. **External backups:**
   Periodically backup `./data/words.db` to external storage

4. **Test rollback procedure:**
   Practice rollback in development before relying on it in production

### Recovery from Corruption

#### SQLite Database Corruption

If database file is corrupted:

1. Try SQLite integrity check:
   ```bash
   sqlite3 ./data/words.db "PRAGMA integrity_check;"
   ```

2. If corrupted, restore from latest backup
3. If no backup available, re-run migration from `dictionary.json`:
   ```bash
   rm ./data/words.db
   dict-gen migrate --input=./cmd/server/dictionary.json
   ```

#### JSON File Corruption

If dictionary.json is corrupted:

1. Restore from backup (see above)
2. OR regenerate from database:
   ```bash
   dict-gen generate --output=./cmd/server/dictionary.json
   ```

## Usage Examples

### Initial Setup (First Time)

1. **Migrate existing dictionary.json to SQLite:**
   ```bash
   # Preview what will be migrated (dry-run)
   dict-gen migrate --input=./cmd/server/dictionary.json --dry-run

   # Perform actual migration
   dict-gen migrate --input=./cmd/server/dictionary.json
   ```

   Output:
   ```
   Starting migration...
      Input: ./cmd/server/dictionary.json
      Database: ./data/words.db

   Backing up existing database...
      Backup created: ./data/words.db.backup.20260127-143022

   Migration complete!
      - 365 words migrated
      - Database: ./data/words.db

    Next steps:
      1. Run: dict-gen validate
      2. Run: dict-gen generate
   ```

2. **Validate the database:**
   ```bash
   dict-gen validate
   ```

3. **Generate dictionary.json:**
   ```bash
   dict-gen generate
   ```

### Adding New Words

**Workflow:**

1. **Add word to SQLite database** (using SQLite CLI or GUI):
   ```bash
   sqlite3 ./data/words.db
   ```

   ```sql
   -- Add word with day_index
   INSERT INTO words (day_index, word, meaning, link, photo, photo_attribution)
   VALUES (42, 'Aroha', 'Love, compassion', 'https://example.com', 'aroha.jpg', 'Photo credit');

   -- Add word without day_index (unassigned pool)
   INSERT INTO words (word, meaning, link, photo, photo_attribution)
   VALUES ('Kōrero', 'Story, discussion', 'https://example.com', 'korero.jpg', 'Photo credit');
   ```

2. **Validate database has 366 unique day indexes:**
   ```bash
   dict-gen validate
   ```

3. **Regenerate dictionary.json from database:**
   ```bash
   dict-gen generate --output=./cmd/server/dictionary.json
   ```

4. **Review changes:**
   ```bash
   git diff ./cmd/server/dictionary.json
   ```

5. **Test server:**
   ```bash
   go run cmd/server/main.go
   ```

6. **Commit changes:**
   ```bash
   git add ./cmd/server/dictionary.json
   git commit -S -m "Add new word: Aroha"
   ```

### Replacing Words for a Specific Day

```bash
# 1. Open database
sqlite3 ./data/words.db

# 2. Find current word for day 42
SELECT * FROM words WHERE day_index = 42;

# 3. Update the word
UPDATE words SET word = 'New Word', meaning = 'New Meaning' WHERE day_index = 42;

# 4. Exit SQLite
.quit

# 5. Regenerate dictionary.json
dict-gen generate

# 6. Validate and commit
dict-gen validate
git add ./cmd/server/dictionary.json
git commit -S -m "Update day 42 word"
```

### Verbose Mode

Get detailed progress information:

```bash
dict-gen --verbose migrate --input=./dictionary.json
```

Output includes:
- File paths being accessed
- Transaction lifecycle events
- Detailed validation results
- Database operation details

### Quiet Mode

Suppress all output except errors (useful for scripts):

```bash
dict-gen --quiet validate
echo $?  # Check exit code: 0 = success, 1 = failure
```

### Compact JSON Output

Generate minified JSON (smaller file size):

```bash
dict-gen generate --compact --output=./cmd/server/dictionary.json
```

### Custom Database Location

```bash
dict-gen migrate --input=./dictionary.json --db=./custom/path/words.db
dict-gen validate --db=./custom/path/words.db
dict-gen generate --db=./custom/path/words.db --output=./output.json
```

### Scheduled Workflow (Cron Example)

```bash
#!/bin/bash
# Daily dictionary regeneration from SQLite

# Regenerate dictionary.json from SQLite
/usr/local/bin/dict-gen generate --output=/app/cmd/server/dictionary.json --quiet

# Restart server if generation succeeded
if [ $? -eq 0 ]; then
    systemctl restart te-reo-bot
fi
```

### Troubleshooting Examples

**Problem: Validation fails with missing indexes**

```bash
$ dict-gen validate

 Validating database...
    Database: ./data/words.db

Validation failed!
   - Total words: 365 (expected 366)
   - Missing indexes: 1
     [42]

 Fix: Ensure all days 1-366 have exactly one word assigned
```

**Solution:**
```bash
sqlite3 ./data/words.db "INSERT INTO words (day_index, word, meaning) VALUES (42, 'Word', 'Meaning');"
dict-gen validate
```

**Problem: Generated JSON is different from source**

This is expected! Generated JSON is sorted by day_index and formatted consistently.

```bash
# Compare meaningful content (not formatting)
diff <(jq -S . original.json) <(jq -S . generated.json)
```

### Integration with Git Workflow

```bash
# Before making changes
git checkout -b update-dictionary

# Make changes to database
sqlite3 ./data/words.db < updates.sql

# Validate
dict-gen validate || exit 1

# Generate
dict-gen generate

# Review
git diff ./cmd/server/dictionary.json

# Commit
git add ./cmd/server/dictionary.json
git commit -S -m "Update dictionary: add new words"

# Push and create PR
git push origin update-dictionary
gh pr create --title "Update dictionary" --body "Added new words to dictionary"
```

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
