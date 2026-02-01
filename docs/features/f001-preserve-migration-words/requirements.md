# Feature Requirements: Preserve Words During Migration

**Related**: [GitHub Issue #11](https://github.com/wizact/te-reo-bot/issues/11) | [Design](./design.md) | [Tasks](./tasks.md)
**Type**: Feature
**Priority**: Medium
**Created**: 2026-01-29

## Overview
Change migration behavior from destructive hard delete to incremental update. Migration will preserve all words in the database, only updating day_index assignments based on dictionary.json. This enables maintaining a word bank separate from daily assignments.

## Requirements

### R0: Deduplication Before Migration

**User Story**: As a CLI developer, I want duplicate words removed from both database and dictionary.json before processing, so that the migration operates on clean, unique data.

**Acceptance Criteria**:
- BEFORE migration processing starts, the system shall remove duplicate word entries from the database, keeping only the first occurrence (lowest ID)
- WHEN processing dictionary.json, the system shall skip duplicate word entries, keeping only the first occurrence (lowest day_index)
- The system shall log the count of duplicates removed from database
- The system shall log the count of duplicates skipped from dictionary.json
- IF duplicates exist in database, THEN result may contain fewer than 366 words, requiring curator intervention
- IF duplicates exist in dictionary.json, THEN migration will assign fewer than 366 day_index values

### R1: Unset Day Index Before Migration

**User Story**: As a CLI developer, I want all day_index values cleared before migration starts, so that the database state is consistent before applying new assignments.

**Acceptance Criteria**:
- WHEN migration begins, the system shall execute `UPDATE words SET day_index = NULL WHERE day_index IS NOT NULL`
- The system shall perform this operation within the same transaction as word imports
- IF unset operation fails, THEN the system shall rollback the transaction and return error

### R2: Match Words by Text

**User Story**: As a CLI developer, I want words matched by exact text comparison, so that existing words are updated rather than duplicated.

**Acceptance Criteria**:
- WHEN processing each dictionary.json entry, the system shall query `SELECT * FROM words WHERE word = ?` using exact case-sensitive match
- The system shall use prepared statements for SQL injection prevention
- The system shall handle both exact matches and non-matches

### R3: Update Existing Words

**User Story**: As a CLI developer, I want existing words updated with new day_index, so that word metadata is preserved across migrations.

**Acceptance Criteria**:
- IF word exists in database, THEN the system shall execute `UPDATE words SET day_index = ? WHERE word = ?`
- The system shall preserve ID, CreatedAt, and IsActive fields during update
- The system shall set UpdatedAt to current timestamp
- WHEN update completes, the system shall log success with word text and new day_index

### R4: Insert New Words

**User Story**: As a CLI developer, I want new words from dictionary.json inserted with day_index, so that the database grows with new entries.

**Acceptance Criteria**:
- IF word does not exist in database, THEN the system shall execute INSERT with all fields from dictionary.json
- The system shall set day_index from dictionary entry
- The system shall auto-generate ID, CreatedAt, UpdatedAt, and set IsActive=true
- WHEN insert completes, the system shall log success with word text and day_index

### R5: Preserve Unmatched Words

**User Story**: As a data curator, I want words in database but not in dictionary.json preserved with NULL day_index, so that I can maintain a word bank for future rotation.

**Acceptance Criteria**:
- WHEN migration completes, the system shall NOT delete words that exist in database but not in dictionary.json
- The system shall leave these words with day_index = NULL after initial unset operation
- The system shall preserve all metadata (ID, Word, Meaning, timestamps, IsActive) for unmatched words
- WHEN validation runs, the system shall allow words with NULL day_index to exist

### R6: Transaction Safety

**User Story**: As a CLI developer, I want all migration operations in a single transaction, so that partial failures do not corrupt the database.

**Acceptance Criteria**:
- WHEN migration starts, the system shall begin transaction before any write operations
- IF any operation fails (unset, update, insert), THEN the system shall rollback entire transaction
- WHEN all operations succeed, the system shall commit transaction atomically
- The system shall log transaction lifecycle events (begin, commit, rollback)

### R7: Migration Logging

**User Story**: As a CLI developer, I want detailed migration logs, so that I can debug issues and track progress.

**Acceptance Criteria**:
- WHEN migration starts, the system shall log total word count from dictionary.json
- The system shall log count of words with day_index before unset operation
- WHEN processing words, the system shall log progress every 50 words (debug level)
- WHEN migration completes, the system shall log success with final word counts
- IF errors occur, THEN the system shall log with stack traces and context (word text, day_index, operation)

### R8: Backward Compatibility

**User Story**: As a CLI developer, I want existing dict-gen commands to work without changes, so that my workflow remains consistent.

**Acceptance Criteria**:
- The system shall maintain existing CLI interface: `dict-gen migrate --input=dictionary.json`
- The system shall maintain existing Repository interface method signatures
- The system shall maintain existing database schema (no migration needed)
- The system shall continue to generate dictionary.json with only words where day_index IS NOT NULL

## Out of Scope

- Soft delete support (is_active flag usage) - Future enhancement
- Word versioning or history tracking
- UI for managing word bank
- Migration from other formats (CSV, Excel)
- Advanced duplicate detection (fuzzy matching, normalization) - Only exact text matches handled

## Success Criteria

1. ✅ All requirements (R0-R8) pass acceptance tests
2. ✅ Duplicate words removed from database before migration
3. ✅ Duplicate words skipped in dictionary.json during processing
4. ✅ Migration preserves words not in dictionary.json
5. ✅ Migration updates existing words without data loss
6. ✅ Migration inserts new words correctly
7. ✅ Transaction rollback works on any failure
8. ✅ Existing tests updated to verify preservation and deduplication behavior
9. ✅ Integration tests confirm incremental migration behavior

## References
- [GitHub Issue #11](https://github.com/wizact/te-reo-bot/issues/11)
- [pkg/migration/migrate.go](https://github.com/wizact/te-reo-bot/blob/master/pkg/migration/migrate.go)
- [pkg/repository/interface.go](https://github.com/wizact/te-reo-bot/blob/master/pkg/repository/interface.go)
- [docs/constitution/tech.md](../../constitution/tech.md)
