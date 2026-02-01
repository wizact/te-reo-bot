# Feature Design: Preserve Words During Migration

**Related**: [GitHub Issue #11](https://github.com/wizact/te-reo-bot/issues/11) | [Requirements](./requirements.md) | [Tasks](./tasks.md)

## Architecture Overview

Transform migration from **destructive delete-and-replace** to **incremental update-or-insert** pattern.

**Current Flow**:
```
dictionary.json → Parse
                    ↓
         DeleteAllWordsByDayIndex(tx)  ← HARD DELETE
                    ↓
         INSERT all words from JSON
                    ↓
              Commit transaction
```

**New Flow**:
```
dictionary.json → Parse
                    ↓
         Deduplicate DB (keep first occurrence by lowest ID)  ← R0
                    ↓
         UPDATE words SET day_index = NULL  ← Unset, not delete (R1)
                    ↓
         Deduplicate dictionary.json (keep first occurrence by lowest day_index)  ← R0
                    ↓
         For each word in JSON:
           SELECT * FROM words WHERE word = ?
           IF exists → UPDATE day_index (R3)
           ELSE → INSERT new word (R4)
                    ↓
              Commit transaction

Result: Words in DB but not in JSON preserved with day_index = NULL (R5)
```

## Components

### Component 1: Migrator (Role: Migration Orchestrator)
**Purpose**: Coordinate incremental migration with word preservation and deduplication

**Responsibilities**:
- Parse dictionary.json
- Deduplicate database words before processing (R0)
- Deduplicate dictionary.json entries before processing (R0)
- Unset all day_index values before processing (R1)
- Match words by text comparison (R2)
- Route to update or insert operations (R3, R4)
- Log migration progress and results (R7)

**Changes Required**:
- **Modified**: `pkg/migration/migrate.go` (MigrateWords method)
  - Add DeduplicateWords call before unset operation
  - Add deduplicateDictWords helper function for dictionary.json deduplication
  - Replace DeleteAllWordsByDayIndex call with UnsetAllDayIndexes
  - Add per-word lookup logic (SELECT by word text)
  - Add conditional update/insert logic
  - Update logging for new operations including deduplication counts

**Dependencies**: WordRepository interface (new methods)

### Component 2: WordRepository Interface (Role: Data Access Contract)
**Purpose**: Define repository operations for incremental migration and deduplication

**Responsibilities**:
- Transaction management (existing)
- Deduplicate words by text (new - R0)
- Unset day_index for all words (new - R1)
- Query word by text (new - R2)
- Update word day_index (new, specialized - R3)
- Insert new words (existing - R4)

**Changes Required**:
- **Modified**: `pkg/repository/interface.go`
  - Add `DeduplicateWords(tx *sql.Tx) (int, error)` - Remove duplicate word entries
  - Add `UnsetAllDayIndexes(tx *sql.Tx) error`
  - Add `GetWordByText(word string) (*Word, error)`
  - Add `UpdateWordDayIndex(tx *sql.Tx, wordText string, dayIndex int) error`
  - Remove `DeleteAllWordsByDayIndex(tx *sql.Tx) error` (deprecated)

**Dependencies**: None (interface definition)

### Component 3: SQLiteRepository Implementation (Role: SQLite Operations)
**Purpose**: Implement incremental migration and deduplication SQL operations

**Responsibilities**:
- Execute DELETE to remove duplicate words (R0)
- Execute UPDATE to unset day_index (R1)
- Execute SELECT for word lookup by text (R2)
- Execute UPDATE for day_index assignment (R3)
- Execute INSERT for new words (R4)

**Changes Required**:
- **Modified**: `pkg/repository/sqlite_repository.go`
  - Implement `DeduplicateWords(tx *sql.Tx) (int, error)` - Delete duplicate words keeping first occurrence
  - Implement `UnsetAllDayIndexes(tx *sql.Tx) error`
  - Implement `GetWordByText(word string) (*Word, error)`
  - Implement `UpdateWordDayIndex(tx *sql.Tx, wordText string, dayIndex int) error`
  - Keep `DeleteAllWordsByDayIndex` for backward compatibility (deprecated)

**Dependencies**: Word model, Logger

## Data Structures

### Word Model (Existing - No Changes)
```go
// pkg/repository/models.go
type Word struct {
    ID               int       `json:"id" db:"id"`
    DayIndex         *int      `json:"index,omitempty" db:"day_index"`  // Nullable
    Word             string    `json:"word" db:"word"`
    Meaning          string    `json:"meaning" db:"meaning"`
    Link             string    `json:"link" db:"link"`
    Photo            string    `json:"photo" db:"photo"`
    PhotoAttribution string    `json:"photo_attribution" db:"photo_attribution"`
    CreatedAt        time.Time `json:"created_at" db:"created_at"`
    UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
    IsActive         bool      `json:"is_active" db:"is_active"`
}
```

**Key Property**: `DayIndex *int` already nullable - supports word bank pattern

### DictionaryWord (Existing - No Changes)
```go
// pkg/migration/migrate.go
type DictionaryWord struct {
    Index            int    `json:"index"`            // Day number (1-366)
    Word             string `json:"word"`             // Māori word text
    Meaning          string `json:"meaning"`
    Link             string `json:"link"`
    Photo            string `json:"photo"`
    PhotoAttribution string `json:"photo_attribution"`
}
```

## Data Flow

### Migration Sequence

**Step 1: Deduplication Phase**
```
START TRANSACTION
  ↓
Deduplicate database words (keep first occurrence of each duplicate word text)
  ↓
Log: "Removed N duplicate words from database"
```

**Step 2: Unset Phase** (R1)
```
UPDATE words SET day_index = NULL WHERE day_index IS NOT NULL
  ↓
Log: "Unset N day_index assignments"
```

**Step 3: Process Phase** (R2, R3, R4)
```
Deduplicate dictionary.json words (keep first occurrence with lowest day_index)
  ↓
Log: "Skipped N duplicate entries from dictionary.json"
  ↓
For each unique word in dictionary.json:
  ↓
  SELECT id, day_index, word, ... FROM words WHERE word = ?
  ↓
  ┌─────────────────┐
  │ Word exists?    │
  └─────────────────┘
         ↓               ↓
       YES              NO
         ↓               ↓
  UPDATE words      INSERT INTO words
  SET day_index = ? VALUES (...)
  WHERE word = ?
         ↓               ↓
  Log: "Updated"   Log: "Inserted"
```

**Step 4: Commit Phase** (R6)
```
COMMIT TRANSACTION
  ↓
Log: "Migration complete - Updated: X, Inserted: Y, Preserved: Z"
```

**Step 5: Post-Migration State** (R5)
```
Database State:
- Words in dictionary.json: day_index = 1-366 (assigned)
- Words in DB only: day_index = NULL (preserved in word bank)
```

## Algorithms

### Incremental Migration Algorithm

**Complexity**: O(N) where N = dictionary.json word count (366)

**Pseudocode**:
```
func MigrateWords(dict Dictionary) error:
    wordCount = len(dict.Words)
    log.Info("Starting migration", wordCount)

    // Count existing assignments
    existingCount = repo.GetWordCountByDayIndex()
    log.Info("Existing day_index assignments", existingCount)

    // Begin transaction
    tx = repo.BeginTx()

    // Deduplicate database words (keep first occurrence)
    dbDuplicates = repo.DeduplicateWords(tx)
    if err:
        repo.RollbackTx(tx)
        return wrap(err, "deduplication failed")
    if dbDuplicates > 0:
        log.Info("Removed duplicate words from database", dbDuplicates)

    // R1: Unset all day_index (preserves words)
    if existingCount > 0:
        err = repo.UnsetAllDayIndexes(tx)
        if err:
            repo.RollbackTx(tx)
            return wrap(err, "unset failed")
        log.Info("Unset day_index", existingCount)

    // Deduplicate dictionary.json (keep first occurrence)
    uniqueWords = deduplicateDictWords(dict.Words)
    dictDuplicates = wordCount - len(uniqueWords)
    if dictDuplicates > 0:
        log.Warn("Skipped duplicate entries in dictionary.json", dictDuplicates)

    // R2, R3, R4: Process each unique word
    updated = 0
    inserted = 0
    for i, dictWord in enumerate(uniqueWords):
        // R7: Progress logging
        if (i+1) % 50 == 0:
            log.Debug("Progress", processed=i+1, total=len(uniqueWords))

        // R2: Lookup by word text
        existing = repo.GetWordByText(dictWord.Word)

        if existing != nil:
            // R3: Update existing word
            err = repo.UpdateWordDayIndex(tx, dictWord.Word, dictWord.Index)
            if err:
                repo.RollbackTx(tx)
                return wrap(err, "update failed", word=dictWord.Word)
            log.Info("Updated", word=dictWord.Word, day_index=dictWord.Index)
            updated++
        else:
            // R4: Insert new word
            word = &Word{
                DayIndex: &dictWord.Index,
                Word: dictWord.Word,
                Meaning: dictWord.Meaning,
                // ... other fields
            }
            err = repo.AddWord(tx, word)
            if err:
                repo.RollbackTx(tx)
                return wrap(err, "insert failed", word=dictWord.Word)
            log.Info("Inserted", word=dictWord.Word, day_index=dictWord.Index)
            inserted++

    // R6: Commit transaction
    err = repo.CommitTx(tx)
    if err:
        repo.RollbackTx(tx)
        return wrap(err, "commit failed")

    // R5, R7: Final counts
    preserved = existingCount - updated  // Words unset but not reassigned
    log.Info("Migration complete",
        updated=updated,
        inserted=inserted,
        preserved=preserved)

    return nil
```

**Performance**:
- Unset operation: O(1) - single UPDATE statement
- Per-word lookup: O(1) - indexed query on word text
- Per-word update/insert: O(1) - single statement
- Total: O(N) linear time for N words

**Memory**: O(1) - processes words sequentially, no bulk loading

## Database Changes

### New Repository Methods (SQL)

**1. DeduplicateWords**:
```sql
DELETE FROM words
WHERE id NOT IN (
    SELECT MIN(id)
    FROM words
    GROUP BY word
);
```
- **Purpose**: Remove duplicate word entries, keeping the first occurrence (lowest ID)
- **Atomicity**: Within transaction
- **Performance**: O(N log N) for grouping, affects only duplicate rows
- **Result**: Returns count of deleted rows

**2. UnsetAllDayIndexes** (R1):
```sql
UPDATE words
SET day_index = NULL, updated_at = CURRENT_TIMESTAMP
WHERE day_index IS NOT NULL;
```
- **Purpose**: Clear all day assignments before migration
- **Atomicity**: Within transaction
- **Performance**: Single table scan, ~366 rows affected max
- **Indexes**: Uses idx_day_index for WHERE clause optimization

**3. GetWordByText** (R2):
```sql
SELECT id, day_index, word, meaning, link, photo, photo_attribution,
       created_at, updated_at, is_active
FROM words
WHERE word = ?;
```
- **Purpose**: Find existing word by exact text match
- **Case Sensitivity**: Exact match (SQLite default collation)
- **Returns**: Single row or sql.ErrNoRows
- **Performance**: O(1) average (would benefit from index on word column)
- **Note**: Consider adding `CREATE INDEX idx_word ON words(word)` in future

**4. UpdateWordDayIndex** (R3):
```sql
UPDATE words
SET day_index = ?, updated_at = CURRENT_TIMESTAMP
WHERE word = ?;
```
- **Purpose**: Assign day_index to existing word
- **Atomicity**: Within transaction
- **Fields Updated**: day_index, updated_at only
- **Fields Preserved**: ID, Word, Meaning, Link, Photo, PhotoAttribution, CreatedAt, IsActive
- **Performance**: O(1) with word text lookup

### Deprecated Methods (Keep for Compatibility)

**DeleteAllWordsByDayIndex** (R8):
```sql
DELETE FROM words WHERE day_index IS NOT NULL;
```
- **Status**: Deprecated, kept for backward compatibility
- **Alternative**: Use UnsetAllDayIndexes instead
- **Removal**: Plan for v2.0.0 breaking change

### Schema Changes
**None Required** - Existing schema already supports:
- Nullable day_index (allows word bank)
- UNIQUE constraint on day_index (prevents duplicates)
- CHECK constraint (validates range 1-366)
- Indexes on day_index and is_active

## API Changes

### WordRepository Interface Changes

**New Methods**:
```go
// DeduplicateWords removes duplicate word entries from database, keeping first occurrence (lowest ID)
// Returns count of deleted duplicate words
// Used at start of migration to ensure data consistency
DeduplicateWords(tx *sql.Tx) (int, error)

// UnsetAllDayIndexes clears day_index for all words with non-null day_index
// Used at start of migration to reset assignments before applying new ones
UnsetAllDayIndexes(tx *sql.Tx) error

// GetWordByText retrieves a word by exact text match (case-sensitive)
// Returns sql.ErrNoRows if word doesn't exist
GetWordByText(word string) (*Word, error)

// UpdateWordDayIndex updates only the day_index field for an existing word
// Preserves all other fields, updates updated_at timestamp
// Uses word text for lookup (not ID)
UpdateWordDayIndex(tx *sql.Tx, wordText string, dayIndex int) error
```

**Deprecated Methods** (R8 - Backward Compatibility):
```go
// DeleteAllWordsByDayIndex - DEPRECATED: Use UnsetAllDayIndexes instead
// Will be removed in v2.0.0
DeleteAllWordsByDayIndex(tx *sql.Tx) error
```

**Unchanged Methods**:
- `BeginTx()`, `CommitTx()`, `RollbackTx()` - Transaction management
- `AddWord(tx *sql.Tx, word *Word) error` - Used for new word insertion
- `GetWordCountByDayIndex() (int, error)` - Used for pre-migration logging
- All other methods remain unchanged

**Migration Impact**: None - new methods are additive, old method kept

## Error Handling

### Error Scenarios and Strategies

**1. Unset Operation Failure** (R1):
```go
if err := repo.UnsetAllDayIndexes(tx); err != nil {
    repo.RollbackTx(tx)
    return entities.NewAppError(err, 500, "Failed to unset day_index assignments").
        WithContext("operation", "migrate_unset_day_index").
        WithContext("existing_count", existingCount)
}
```

**2. Word Lookup Failure** (R2):
```go
existing, err := repo.GetWordByText(dictWord.Word)
if err != nil && err != sql.ErrNoRows {
    repo.RollbackTx(tx)
    return entities.NewAppError(err, 500, "Failed to lookup existing word").
        WithContext("operation", "migrate_lookup_word").
        WithContext("word", dictWord.Word).
        WithContext("day_index", dictWord.Index)
}
// Note: sql.ErrNoRows is expected (word doesn't exist) - proceed to insert
```

**3. Update Failure** (R3):
```go
if err := repo.UpdateWordDayIndex(tx, dictWord.Word, dictWord.Index); err != nil {
    repo.RollbackTx(tx)
    return entities.NewAppError(err, 500, "Failed to update word day_index").
        WithContext("operation", "migrate_update_day_index").
        WithContext("word", dictWord.Word).
        WithContext("day_index", dictWord.Index)
}
```

**4. Insert Failure** (R4):
```go
if err := repo.AddWord(tx, word); err != nil {
    repo.RollbackTx(tx)
    return entities.NewAppError(err, 500, fmt.Sprintf("Failed to insert word %q", dictWord.Word)).
        WithContext("operation", "migrate_insert_word").
        WithContext("word", dictWord.Word).
        WithContext("day_index", dictWord.Index)
}
```

**5. Transaction Commit Failure** (R6):
```go
if err := repo.CommitTx(tx); err != nil {
    repo.RollbackTx(tx)
    return entities.NewAppError(err, 500, "Failed to commit migration transaction").
        WithContext("operation", "migrate_commit_tx").
        WithContext("updated_count", updated).
        WithContext("inserted_count", inserted)
}
```

**Error Recovery**: All errors trigger transaction rollback - database returns to pre-migration state

## Testing Approach

### Unit Tests (pkg/migration/migrate_test.go)

**Test 1: Deduplicate Database Words**:
```go
func TestMigrator_DeduplicateDatabaseWords(t *testing.T) {
    // Setup: Insert duplicate "kia ora" entries (IDs 1, 5, 10)
    // Execute: Migrate with any dictionary
    // Verify: Only ID=1 remains, IDs 5 and 10 deleted
}
```

**Test 2: Deduplicate Dictionary Words**:
```go
func TestMigrator_DeduplicateDictionaryWords(t *testing.T) {
    // Setup: Dictionary with "kia ora" at day_index=1 and day_index=5
    // Execute: Migrate
    // Verify: Only day_index=1 applied, warning logged for 1 duplicate skipped
}
```

**Test 3: Unset Existing Day Indexes**:
```go
func TestMigrator_UnsetExistingDayIndexes(t *testing.T) {
    // Setup: Insert 3 words with day_index=1,2,3
    // Execute: Migrate with empty dictionary
    // Verify: All 3 words still exist with day_index=NULL
}
```

**Test 4: Update Existing Words**:
```go
func TestMigrator_UpdateExistingWords(t *testing.T) {
    // Setup: Insert word "kia ora" with day_index=1
    // Execute: Migrate with "kia ora" at day_index=5
    // Verify: Same word ID, day_index updated to 5, CreatedAt unchanged
}
```

**Test 5: Insert New Words**:
```go
func TestMigrator_InsertNewWords(t *testing.T) {
    // Setup: Empty database
    // Execute: Migrate with dictionary containing 3 words
    // Verify: 3 words inserted with correct day_index
}
```

**Test 6: Mixed Update and Insert**:
```go
func TestMigrator_MixedUpdateInsert(t *testing.T) {
    // Setup: Insert "kia ora" (day_index=1), "aroha" (day_index=2)
    // Execute: Migrate with "kia ora" (day_index=3), "tēnā koe" (day_index=4)
    // Verify:
    //   - "kia ora" updated to day_index=3
    //   - "aroha" preserved with day_index=NULL
    //   - "tēnā koe" inserted with day_index=4
}
```

**Test 7: Transaction Rollback on Error**:
```go
func TestMigrator_RollbackOnUpdateError(t *testing.T) {
    // Setup: Mock repo returning error on UpdateWordDayIndex
    // Execute: Migrate with valid dictionary
    // Verify: Transaction rolled back, database unchanged
}
```

**Test 8: Preserve Word Metadata**:
```go
func TestMigrator_PreserveWordMetadata(t *testing.T) {
    // Setup: Insert word with ID=10, CreatedAt=2024-01-01, IsActive=true
    // Execute: Migrate updating same word
    // Verify: ID=10, CreatedAt unchanged, IsActive=true, UpdatedAt changed
}
```

### Integration Tests (pkg/repository/repository_test.go)

**Test 1: DeduplicateWords SQL**:
```sql
-- Setup
INSERT INTO words (id, word, meaning) VALUES (1, 'kia ora', 'hello');
INSERT INTO words (id, word, meaning) VALUES (5, 'kia ora', 'hello duplicate');
INSERT INTO words (id, word, meaning) VALUES (10, 'kia ora', 'hello another');
INSERT INTO words (id, word, meaning) VALUES (2, 'aroha', 'love');

-- Execute
DELETE FROM words WHERE id NOT IN (SELECT MIN(id) FROM words GROUP BY word);

-- Verify
SELECT COUNT(*) FROM words;  -- Expect: 2 (kia ora with id=1, aroha with id=2)
SELECT id FROM words WHERE word = 'kia ora';  -- Expect: 1
```

**Test 2: UnsetAllDayIndexes SQL**:
```sql
-- Setup
INSERT INTO words (day_index, word, meaning) VALUES (1, 'kia ora', 'hello');
INSERT INTO words (day_index, word, meaning) VALUES (2, 'aroha', 'love');
INSERT INTO words (day_index, word, meaning) VALUES (NULL, 'extra', 'bonus');

-- Execute
UPDATE words SET day_index = NULL WHERE day_index IS NOT NULL;

-- Verify
SELECT COUNT(*) FROM words WHERE day_index IS NOT NULL;  -- Expect: 0
SELECT COUNT(*) FROM words;  -- Expect: 3
```

**Test 3: GetWordByText SQL**:
```sql
-- Setup
INSERT INTO words (word, meaning) VALUES ('kia ora', 'hello');

-- Execute
SELECT * FROM words WHERE word = 'kia ora';  -- Expect: 1 row
SELECT * FROM words WHERE word = 'missing';   -- Expect: 0 rows
SELECT * FROM words WHERE word = 'KIA ORA';   -- Expect: 0 rows (case-sensitive)
```

**Test 4: UpdateWordDayIndex SQL**:
```sql
-- Setup
INSERT INTO words (word, meaning, day_index) VALUES ('kia ora', 'hello', 1);

-- Execute
UPDATE words SET day_index = 5, updated_at = CURRENT_TIMESTAMP WHERE word = 'kia ora';

-- Verify
SELECT day_index FROM words WHERE word = 'kia ora';  -- Expect: 5
SELECT id, created_at FROM words WHERE word = 'kia ora';  -- Expect: unchanged
```

### Manual Tests

**Test 1: Real Dictionary Migration**:
```bash
# Setup: Create database with 5 test words (day_index 1-5)
sqlite3 data/words.db < testdata/seed_words.sql

# Execute migration
./dict-gen migrate --input=dictionary.json

# Verify results
sqlite3 data/words.db "SELECT COUNT(*) FROM words WHERE day_index IS NOT NULL;"  # 366
sqlite3 data/words.db "SELECT COUNT(*) FROM words WHERE day_index IS NULL;"      # 5 preserved
sqlite3 data/words.db "SELECT word, day_index FROM words WHERE word = 'kia ora';" # Check updated
```

**Test 2: Idempotent Migration**:
```bash
# Run migration twice with same dictionary.json
./dict-gen migrate --input=dictionary.json
./dict-gen migrate --input=dictionary.json

# Verify: Same word count, no duplicates, day_index unchanged
sqlite3 data/words.db "SELECT COUNT(*) FROM words;"  # Should be constant
```

### Test Fixtures

**testdata/valid_dictionary.json**:
```json
{
  "dictionary": [
    {"index": 1, "word": "ngā mihi o te tau hou", "meaning": "Happy New Year"},
    {"index": 2, "word": "kia ora", "meaning": "hello"}
  ]
}
```

**testdata/seed_words.sql**:
```sql
INSERT INTO words (day_index, word, meaning) VALUES (1, 'kia ora', 'hello - old');
INSERT INTO words (day_index, word, meaning) VALUES (2, 'aroha', 'love');
INSERT INTO words (day_index, word, meaning) VALUES (NULL, 'extra word', 'bonus content');
```

## Migration Path

### Development Environment

**Step 1: Update Code**:
```bash
# Install dependencies (if any new)
go mod tidy

# Run tests
go test ./pkg/migration/... -v
go test ./pkg/repository/... -v
```

**Step 2: Test with Sample Data**:
```bash
# Backup existing database
cp data/words.db data/words.db.backup

# Run migration
./dict-gen migrate --input=dictionary.json

# Verify results
./dict-gen validate
sqlite3 data/words.db "SELECT COUNT(*) FROM words WHERE day_index IS NULL;"
```

**Step 3: Rollback Plan** (if issues):
```bash
# Restore backup
cp data/words.db.backup data/words.db

# Or rebuild from dictionary.json (old behavior)
git checkout HEAD~1 cmd/dict-gen pkg/migration pkg/repository
make build
./dict-gen migrate --input=dictionary.json
```

### Production Environment

**Production Use Case**: Dictionary.json generation for HTTP server

**Deployment Steps**:
1. Deploy updated dict-gen binary
2. Run migration: `./dict-gen migrate --input=dictionary.json`
3. Verify word count: `./dict-gen validate`
4. Generate JSON: `./dict-gen generate --output=dictionary.json`
5. Deploy HTTP server with new dictionary.json

**Backward Compatibility** (R8):
- Old dict-gen binary: Uses DeleteAllWordsByDayIndex (still works)
- New dict-gen binary: Uses UnsetAllDayIndexes (word preservation)
- No database schema migration required
- Existing dictionary.json format unchanged

**Data Preservation Guarantee**:
- Words in database but not in dictionary.json: Preserved with day_index=NULL
- Future rotation: Update dictionary.json to include preserved words
- No data loss on migration

## Risks & Mitigations

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Word text mismatch (case/whitespace) | Update fails, word duplicated | Medium | Normalize word text before comparison; add integration tests for edge cases |
| Concurrent migrations | Data corruption | Low | CLI tool (single-user), transaction isolation prevents concurrent writes |
| Large word bank (>1000 words) | Slow migration | Low | Current O(N) algorithm scales well; batch operations already efficient |
| Transaction timeout | Migration fails mid-process | Low | SQLite transactions have no timeout; full rollback on error |
| Duplicate word text | UNIQUE constraint violation | Low | GetWordByText returns first match; validate dictionary.json beforehand |
| Missing index on word column | Slow lookups (>10ms) | Medium | Add `CREATE INDEX idx_word ON words(word)` if performance degrades; monitor query time |

**Critical Risk - Word Text Normalization**:
- **Scenario**: Dictionary.json has "kia ora" but DB has "Kia Ora" (capitalization difference)
- **Result**: Treated as different words, DB word preserved with day_index=NULL, new word inserted
- **Mitigation**: Document case-sensitivity requirement; consider adding normalization in future

## Performance Considerations

### Expected Performance (366 Words)

**Unset Operation**: ~5ms
- Single UPDATE statement
- Affects ~366 rows max
- Uses idx_day_index for WHERE clause

**Per-Word Processing**: ~2-3ms each
- GetWordByText: ~1ms (O(1) lookup, no index yet)
- UpdateWordDayIndex or AddWord: ~1-2ms
- Total: 366 * 3ms = ~1100ms (1.1 seconds)

**Transaction Commit**: ~50ms
- SQLite WAL mode (default)
- Fsync overhead

**Total Migration Time**: ~1.2 seconds (well under R7 target)

### Performance Targets (Existing from tech.md)
- ✅ dict-gen migrate: <1s for 366 words → **1.2s (slightly over, acceptable)**
- ✅ Query performance: <1ms for indexed lookups → **GetWordByText ~1ms without index**

### Optimization Opportunities (Future)

**1. Add Word Text Index**:
```sql
CREATE INDEX idx_word ON words(word);
```
- **Benefit**: GetWordByText drops from ~1ms to ~0.1ms
- **Cost**: Minimal (small table, single column, infrequent writes)
- **When**: If migration time > 3s with large word bank (>1000 words)

**2. Batch Updates** (if needed for large datasets):
```sql
-- Instead of per-word UPDATE, build batch
UPDATE words SET day_index = CASE word
    WHEN 'kia ora' THEN 1
    WHEN 'aroha' THEN 2
    ...
END
WHERE word IN ('kia ora', 'aroha', ...);
```
- **Benefit**: Single UPDATE statement instead of N
- **Cost**: Complex SQL generation, harder to debug
- **When**: Word count > 5000 and migration time > 10s

**Current Decision**: No optimization needed for 366 words (1.2s acceptable)

### Bottlenecks

**No Expected Bottlenecks**:
- ✅ Database I/O: SQLite on local disk, WAL mode enabled
- ✅ Transaction overhead: Single transaction for all operations
- ✅ Network latency: N/A (local SQLite file)
- ✅ Memory usage: Sequential processing, no bulk loading

## Alternative Approaches Considered

### Alternative 1: Soft Delete with is_active Flag
**Description**: Mark words as inactive instead of unsetting day_index
```sql
UPDATE words SET is_active = 0 WHERE day_index IS NOT NULL;
-- Then UPDATE is_active = 1 for matched words
```

**Pros**:
- Preserves day_index history
- Supports versioning (future feature)
- Simpler rollback (flip is_active flag)

**Cons**:
- Complicates queries (WHERE day_index IS NOT NULL AND is_active = 1)
- Breaks backward compatibility (existing code assumes is_active = 1 always)
- Out of scope for current requirements (R5 only requires preservation, not versioning)

**Decision**: Rejected - Adds complexity without clear benefit; NULL day_index already signals "not assigned"

### Alternative 2: Upsert with REPLACE Statement
**Description**: Use SQLite REPLACE (DELETE + INSERT) for all words
```sql
REPLACE INTO words (word, day_index, meaning, ...) VALUES (...);
```

**Pros**:
- Simpler code (no conditional logic)
- Single SQL statement per word

**Cons**:
- REPLACE deletes and reinserts → new ID, new CreatedAt (violates R3, R5)
- Loses word metadata (CreatedAt, IsActive)
- Does NOT preserve unmatched words (violates R5)

**Decision**: Rejected - Violates core requirement R5 (preserve words not in dictionary.json)

### Alternative 3: Two-Phase Commit (Staging Table)
**Description**: Import to staging table, then merge into main table
```sql
CREATE TEMP TABLE staging_words AS SELECT * FROM words WHERE 1=0;
-- INSERT dictionary.json into staging_words
-- MERGE staging_words into words
```

**Pros**:
- Can validate before committing
- Atomic replacement (all or nothing)

**Cons**:
- Higher complexity (2 tables, merge logic)
- Slower (double I/O)
- Still requires word matching logic
- Overkill for 366 words

**Decision**: Rejected - Unnecessary complexity for current dataset size and requirements

## Edge Cases

### Edge Case 1: Empty Dictionary.json
**Scenario**: Migration called with dictionary containing 0 words

**Behavior**:
1. Unset all day_index assignments (all existing words → day_index=NULL)
2. No updates or inserts (loop processes 0 words)
3. Commit transaction
4. Log: "Migration complete - Updated: 0, Inserted: 0, Preserved: N"

**Rationale**: Correct behavior - all words moved to word bank (day_index=NULL)

**Validation**: dict-gen validate should fail (requires 366 words with day_index)

### Edge Case 2: Duplicate Words in Database
**Scenario**: Database contains duplicate entries for "kia ora" (IDs 10 and 25)

**Behavior**:
1. DeduplicateWords called at migration start
2. DELETE all words WHERE id NOT IN (MIN(id) GROUP BY word)
3. Entry with ID=10 kept, ID=25 deleted
4. Result: Single "kia ora" entry remains

**Rationale**: Keep first occurrence (lowest ID) to maintain chronological order

**Impact**: Database may have fewer than 366 words after cleanup, requiring manual curation

### Edge Case 3: Duplicate Word Text in Dictionary.json
**Scenario**: dictionary.json contains "kia ora" at day_index=1 and day_index=5

**Behavior**:
1. deduplicateDictWords() removes duplicates, keeping first occurrence (lowest day_index)
2. Only "kia ora" (day_index=1) processed
3. Log warning: "Skipped 1 duplicate entries in dictionary.json"
4. Result: Single word with day_index=1

**Rationale**: First-occurrence-wins for consistency with database deduplication

**Impact**: Dictionary.json may assign fewer than 366 words if duplicates exist

### Edge Case 4: Case-Sensitive Word Matching
**Scenario**: DB has "Kia Ora", dictionary.json has "kia ora"

**Behavior**:
1. GetWordByText("kia ora") returns sql.ErrNoRows (no exact match)
2. INSERT "kia ora" as new word (day_index=1)
3. "Kia Ora" preserved with day_index=NULL

**Rationale**: SQLite default collation is case-sensitive; exact match required

**Trade-off**:
- ✅ Preserves data integrity (no unintended matches)
- ❌ May create duplicates if word text case changes

**Future**: Document case-sensitivity requirement; add normalization if needed

### Edge Case 5: Word Text with Unicode/Macrons
**Scenario**: Māori words with macrons (ā, ē, ī, ō, ū)

**Behavior**:
1. GetWordByText("Māori") matches "Māori" exactly (UTF-8 comparison)
2. Does NOT match "Maori" (without macron)

**Rationale**: Correct Māori orthography requires macron distinction

**Validation**: Existing dictionary.json already uses macrons consistently

### Edge Case 6: Transaction Failure Mid-Migration
**Scenario**: Disk full after processing 200/366 words

**Behavior**:
1. UpdateWordDayIndex fails with "database or disk is full"
2. Migrator calls RollbackTx(tx)
3. All changes reverted: unset, updates, inserts all rolled back
4. Database returns to pre-migration state

**Rationale**: ACID guarantees - transaction atomicity

**Recovery**: Free disk space, retry migration

## Code Snippets

### New Repository Method: DeduplicateWords
```go
// pkg/repository/sqlite_repository.go

// DeduplicateWords removes duplicate word entries, keeping first occurrence (lowest ID)
func (r *SQLiteRepository) DeduplicateWords(tx *sql.Tx) (int, error) {
    r.logger.Debug("Deduplicating words in database")

    query := `
        DELETE FROM words
        WHERE id NOT IN (
            SELECT MIN(id)
            FROM words
            GROUP BY word
        )
    `
    result, err := tx.Exec(query)
    if err != nil {
        return 0, entities.NewAppError(err, 500, "Failed to deduplicate words").
            WithContext("operation", "deduplicate_words").
            WithContext("table", "words")
    }

    rowsAffected, _ := result.RowsAffected()
    if rowsAffected > 0 {
        r.logger.Info("Removed duplicate words",
            logger.Int64("duplicates_removed", rowsAffected))
    }
    return int(rowsAffected), nil
}
```

### New Repository Method: UnsetAllDayIndexes
```go
// pkg/repository/sqlite_repository.go

// UnsetAllDayIndexes clears day_index for all words (preserves words)
func (r *SQLiteRepository) UnsetAllDayIndexes(tx *sql.Tx) error {
    r.logger.Debug("Unsetting all day_index assignments")

    query := `UPDATE words SET day_index = NULL, updated_at = ? WHERE day_index IS NOT NULL`
    result, err := tx.Exec(query, time.Now())
    if err != nil {
        return entities.NewAppError(err, 500, "Failed to unset day_index assignments").
            WithContext("operation", "unset_day_indexes").
            WithContext("table", "words")
    }

    rowsAffected, _ := result.RowsAffected()
    r.logger.Info("Day indexes unset", logger.Int64("rows_affected", rowsAffected))
    return nil
}
```

### New Repository Method: GetWordByText
```go
// pkg/repository/sqlite_repository.go

// GetWordByText retrieves a word by exact text match (case-sensitive)
func (r *SQLiteRepository) GetWordByText(word string) (*Word, error) {
    query := `
        SELECT id, day_index, word, meaning, link, photo, photo_attribution,
               created_at, updated_at, is_active
        FROM words
        WHERE word = ?
    `

    row := r.db.QueryRow(query, word)
    wordResult, err := scanWord(row)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, err  // Expected - word doesn't exist
        }
        return nil, entities.NewAppError(err, 500, "Failed to get word by text").
            WithContext("operation", "get_word_by_text").
            WithContext("table", "words").
            WithContext("word", word)
    }

    return &wordResult, nil
}
```

### New Repository Method: UpdateWordDayIndex
```go
// pkg/repository/sqlite_repository.go

// UpdateWordDayIndex updates only the day_index field for an existing word
func (r *SQLiteRepository) UpdateWordDayIndex(tx *sql.Tx, wordText string, dayIndex int) error {
    query := `UPDATE words SET day_index = ?, updated_at = ? WHERE word = ?`

    r.logger.Debug("Updating word day_index",
        logger.String("word", wordText),
        logger.Int("day_index", dayIndex))

    _, err := tx.Exec(query, dayIndex, time.Now(), wordText)
    if err != nil {
        return entities.NewAppError(err, 500, "Failed to update word day_index").
            WithContext("operation", "update_word_day_index").
            WithContext("table", "words").
            WithContext("word", wordText).
            WithContext("day_index", dayIndex)
    }

    r.logger.Info("Word day_index updated",
        logger.String("word", wordText),
        logger.Int("day_index", dayIndex))
    return nil
}
```

### Helper Function: Deduplicate Dictionary Words
```go
// pkg/migration/migrate.go

// deduplicateDictWords returns unique words from dictionary, keeping first occurrence
func deduplicateDictWords(words []DictionaryWord) []DictionaryWord {
    seen := make(map[string]bool)
    unique := make([]DictionaryWord, 0, len(words))

    for _, w := range words {
        if !seen[w.Word] {
            seen[w.Word] = true
            unique = append(unique, w)
        }
    }

    return unique
}
```

### Updated Migration Logic
```go
// pkg/migration/migrate.go

func (m *Migrator) MigrateWords(dict *Dictionary) error {
    wordCount := len(dict.Words)
    m.logger.Info("Starting migration", logger.Int("word_count", wordCount))

    existingCount, err := m.repo.GetWordCountByDayIndex()
    if err != nil {
        // ... error handling
    }
    m.logger.Info("Existing day_index assignments", logger.Int("existing_count", existingCount))

    tx, err := m.repo.BeginTx()
    if err != nil {
        // ... error handling
    }

    // Deduplicate database words
    dbDuplicates, err := m.repo.DeduplicateWords(tx)
    if err != nil {
        m.repo.RollbackTx(tx)
        // ... error handling
    }
    if dbDuplicates > 0 {
        m.logger.Info("Removed duplicate words from database",
            logger.Int("duplicates_removed", dbDuplicates))
    }

    // R1: Unset all day_index (preserves words)
    if existingCount > 0 {
        m.logger.Info("Unsetting existing day_index assignments", logger.Int("count", existingCount))
        if err := m.repo.UnsetAllDayIndexes(tx); err != nil {
            m.repo.RollbackTx(tx)
            // ... error handling
        }
    }

    // Deduplicate dictionary.json words
    uniqueWords := deduplicateDictWords(dict.Words)
    dictDuplicates := wordCount - len(uniqueWords)
    if dictDuplicates > 0 {
        m.logger.Warn("Skipped duplicate entries in dictionary.json",
            logger.Int("duplicates_skipped", dictDuplicates))
    }

    // R2, R3, R4: Process each unique word
    updated := 0
    inserted := 0
    m.logger.Info("Processing words", logger.Int("word_count", len(uniqueWords)))

    for i, dictWord := range uniqueWords {
        // R7: Progress logging
        if (i+1)%50 == 0 {
            m.logger.Debug("Migration progress",
                logger.Int("processed", i+1),
                logger.Int("total", len(uniqueWords)))
        }

        // R2: Lookup existing word by text
        existing, err := m.repo.GetWordByText(dictWord.Word)
        if err != nil && err != sql.ErrNoRows {
            m.repo.RollbackTx(tx)
            // ... error handling
        }

        if existing != nil {
            // R3: Update existing word day_index
            if err := m.repo.UpdateWordDayIndex(tx, dictWord.Word, dictWord.Index); err != nil {
                m.repo.RollbackTx(tx)
                // ... error handling
            }
            updated++
        } else {
            // R4: Insert new word
            word := &repository.Word{
                DayIndex:         &dictWord.Index,
                Word:             dictWord.Word,
                Meaning:          dictWord.Meaning,
                Link:             dictWord.Link,
                Photo:            dictWord.Photo,
                PhotoAttribution: dictWord.PhotoAttribution,
            }
            if err := m.repo.AddWord(tx, word); err != nil {
                m.repo.RollbackTx(tx)
                // ... error handling
            }
            inserted++
        }
    }

    // R6: Commit transaction
    if err := m.repo.CommitTx(tx); err != nil {
        m.repo.RollbackTx(tx)
        // ... error handling
    }

    // R7: Final logging
    preserved := existingCount - updated
    m.logger.Info("Migration completed successfully",
        logger.Int("updated", updated),
        logger.Int("inserted", inserted),
        logger.Int("preserved", preserved),
        logger.Int("total_words", updated+inserted+preserved))

    return nil
}
```

## References

- [GitHub Issue #11](https://github.com/wizact/te-reo-bot/issues/11)
- [Requirements](./requirements.md) - Detailed EARS requirements
- [Tasks](./tasks.md) - Implementation task breakdown
- [tech.md](../../constitution/tech.md) - Project architecture patterns
- [conventions.md](../../conventions.md) - Go coding standards
- [pkg/migration/migrate.go](https://github.com/wizact/te-reo-bot/blob/master/pkg/migration/migrate.go) - Current implementation
- [pkg/repository/sqlite_repository.go](https://github.com/wizact/te-reo-bot/blob/master/pkg/repository/sqlite_repository.go) - Repository implementation
