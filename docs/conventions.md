# Go Code Conventions - Te Reo Bot

This document establishes coding standards for the te-reo-bot project.

## Code Conventions

### File Naming

**All lowercase** with underscores separating words:
- ✅ `word_selector.go`, `http_server.go`, `google_cloud_storage.go`
- ❌ `wordSelector.go`, `http-server.go`, `HttpServer.go`

**Test files** append `_test.go`:
- ✅ `word_selector_test.go`, `integration_test.go`
- ❌ `word_selector_tests.go`, `test_word_selector.go`

**Benchmark files** append `_bench_test.go`:
- ✅ `migrate_bench_test.go`, `generator_bench_test.go`

### Package Naming

**Lowercase, single words** matching directory names:
- ✅ `wotd`, `logger`, `repository`, `validator`
- ❌ `word_of_the_day`, `repo`, `utils`, `helpers`

Avoid generic names like `common`, `util`, `helper`. Use descriptive, domain-specific names.

### Naming Conventions

**Interfaces**: Descriptive names with conventional suffixes (`-er`, `-or`):
- ✅ `Logger`, `WordRepository`, `WordSelector`
- ❌ `ILogger`, `LoggerInterface`, `AbstractLogger`

**Structs**: PascalCase for exported, camelCase for unexported:
- ✅ `WordSelector`, `AppError`, `sqliteRepository`
- ❌ `word_selector`, `appError`

**Variables**: camelCase:
- ✅ `wordCount`, `dayIndex`, `apiKey`
- ❌ `word_count`, `DayIndex`, `API_KEY`

**Constants**: PascalCase or SCREAMING_SNAKE_CASE for package-level:
- ✅ `MaxRetries`, `DEFAULT_PORT`, `minWordLength`
- ❌ `maxRetries`, `default_port`

**Error variables**: `err` for temporary, `Err` prefix for package-level:
```go
err := DoSomething()
var ErrNotFound = errors.New("not found")
```

## Testing Conventions

### Framework

**stdlib testing + testify**:
```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)
```

**Use `require` for critical assertions** (stops test execution):
```go
db, err := setupTestDB(t)
require.NoError(t, err)  // Must succeed to continue
```

**Use `assert` for soft failures** (test continues):
```go
assert.Equal(t, 366, wordCount)
assert.NotEmpty(t, word.Meaning)
```

### Table-Driven Tests

**Structure** with subtests:
```go
func TestWordSelector_SelectWordByDay(t *testing.T) {
    tests := []struct {
        name     string
        day      int
        expected string
        wantErr  bool
    }{
        {"day 1", 1, "ngā mihi o te tau hou", false},
        {"day 366 leap year", 366, "Hōngongoi", false},
        {"invalid day 0", 0, "", true},
        {"invalid day 367", 367, "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            word, err := selector.SelectWordByDay(tt.day)
            if tt.wantErr {
                require.Error(t, err)
                return
            }
            require.NoError(t, err)
            assert.Equal(t, tt.expected, word.Word)
        })
    }
}
```

### Test Fixtures

**testdata/ directories** for sample data:
```
pkg/
  validator/
    validator.go
    validator_test.go
    testdata/
      valid_dictionary.json
      invalid_missing_day.json
      invalid_duplicate_day.json
```

### Helper Functions

**Reduce repetition** across tests:
```go
func setupTestDB(t *testing.T) *sql.DB {
    t.Helper()
    db, err := sql.Open("sqlite3", ":memory:")
    require.NoError(t, err)
    _, err = db.Exec(schema)
    require.NoError(t, err)
    return db
}
```

### Integration Tests

**Separate files** with build tags (optional):
```go
// +build integration

package wotd

import "testing"

func TestMastodonClient_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    // ...
}
```

Run with: `go test -tags=integration`

## Error Handling

### Basic Principles

**Check errors immediately**:
```go
// ✅ Good
file, err := os.Open("file.txt")
if err != nil {
    return fmt.Errorf("failed to open file: %w", err)
}
defer file.Close()

// ❌ Bad
file, _ := os.Open("file.txt")  // Ignoring error
defer file.Close()
```

**Wrap errors with context** using `%w`:
```go
// ✅ Good - enables errors.Is() and errors.As()
err := parser.Parse(data)
if err != nil {
    return fmt.Errorf("failed to parse dictionary: %w", err)
}

// ❌ Bad - breaks error chain
return fmt.Errorf("failed to parse dictionary: %v", err)
```

**Don't log and return** (choose one):
```go
// ✅ Good - return error, let caller decide to log
if err := db.Execute(); err != nil {
    return fmt.Errorf("database error: %w", err)
}

// ❌ Bad - logs AND returns (double logging)
if err != nil {
    logger.Error(err, "database error")
    return err
}
```

**Error messages**: lowercase, no ending punctuation:
```go
// ✅ Good
errors.New("database connection failed")
fmt.Errorf("failed to parse word %q", word)

// ❌ Bad
errors.New("Database connection failed.")
fmt.Errorf("Failed to parse word: %s!", word)
```

### Custom AppError Type

**For HTTP handlers** (`pkg/entities/http-entities.go`):
```go
type AppError struct {
    Err        error                  // Original error (json:"-")
    Message    string                 // User-facing message
    Code       int                    // HTTP status code
    StackTrace *logger.StackTrace     // Captured stack trace (json:"-")
    Context    map[string]interface{} // Debug metadata
}
```

**Usage**:
```go
appErr := entities.NewAppError(err, 500, "Failed to upload file")
appErr.WithContext("word", wordName)
appErr.WithContext("operation", "media_upload")

logger.ErrorWithStack(appErr, "Media upload failed",
    logger.String("word", wordName),
    logger.String("operation", "media_upload"),
)
return appErr
```

**Security**: Context and stack traces never exposed in HTTP responses.

### Three-Tier Error Handling

| Layer | Pattern | Tool | Exit Strategy |
|-------|---------|------|---------------|
| **Library/Package** | Error wrapping | `fmt.Errorf("%w")` | Return error |
| **HTTP Handler** | AppError with context | Custom type | HTTP response |
| **Middleware** | Panic recovery + logging | defer + logger | HTTP 500 |
| **Server Start** | Fatal logging | `logger.Fatal()` | `os.Exit(1)` |
| **CLI** | Simple fmt printf | fmt.Fprintf | `os.Exit(1)` |
| **Database** | Transaction rollback | Explicit rollback | Return error |

### Transaction Rollback Pattern

**Always rollback on error**:
```go
tx, err := repo.BeginTx()
if err != nil {
    return fmt.Errorf("failed to begin transaction: %w", err)
}

if err := repo.AddWord(tx, word); err != nil {
    repo.RollbackTx(tx)  // Always rollback
    return fmt.Errorf("failed to add word %q: %w", word.Word, err)
}

if err := repo.CommitTx(tx); err != nil {
    repo.RollbackTx(tx)  // Rollback if commit fails
    return fmt.Errorf("failed to commit transaction: %w", err)
}
```

## Structured Logging

### Custom Logger (pkg/logger)

**Custom implementation** with zero external dependencies:
- JSON and text format support
- Stack trace capture
- Request context tracking
- Environment-aware configuration

**Logger Interface**:
```go
type Logger interface {
    Error(err error, message string, fields ...Field)
    ErrorWithStack(err error, message string, fields ...Field)
    Fatal(err error, message string, fields ...Field)
    Info(message string, fields ...Field)
    Debug(message string, fields ...Field)
}
```

### Log Levels

- `debug` - Detailed diagnostic information
- `info` - General informational messages (default)
- `error` - Error conditions
- `fatal` - Fatal errors (`os.Exit(1)`)

### Structured Fields

**Type-safe field constructors**:
```go
logger.Info("Server starting",
    logger.String("version", version.GetVersion()),
    logger.String("git_commit", version.GetGitCommit()),
    logger.Int("port", 8080),
    logger.Bool("tls_enabled", true),
)
```

**Never use string formatting** in log messages:
```go
// ✅ Good - structured, queryable
logger.Error(err, "Failed to upload media",
    logger.String("word", wordName),
    logger.String("bucket", bucketName),
)

// ❌ Bad - unstructured, hard to query
logger.Error(err, fmt.Sprintf("Failed to upload media for word %s to bucket %s", wordName, bucketName))
```

### Configuration

**Environment Variables**:
- `ENABLE_STACK_TRACES` (default: `true`)
- `LOG_LEVEL` (default: `info`, values: debug/info/error/fatal)
- `ENVIRONMENT` (default: `dev`, values: dev/prod/test)
- `LOG_FORMAT` (default: `json`, values: json/text)

**Global Logger Pattern**:
```go
// Initialize early in main()
logger.InitializeGlobalLogger(nil)

// Access anywhere
log := logger.GetGlobalLogger()
log.Info("Operation started")
```

### Security - Stack Traces

**Stack traces NEVER in HTTP responses**:
- Stored in `AppError.StackTrace` field tagged `json:"-"`
- Only logged server-side
- HTTP returns `FriendlyError` with generic messages

**When to use**:
- `ErrorWithStack()`: Unexpected errors, need debugging
- `Error()`: Expected/handled errors
- `Fatal()`: Always includes stack trace

### Request Context Tracking

**Automatic HTTP request tracking**:
```go
reqCtx := logger.ExtractRequestContext(r)
fields := reqCtx.ToFields()

logger.Info("Request received", fields...)
```

**Captured Fields**:
- `request_id`: Unique identifier (16-char hex)
- `request_method`: HTTP method
- `request_path`: Request path
- `request_user_agent`: User agent
- `request_remote_addr`: Client IP (handles proxy headers)

### Logging Best Practices

**Do Log**:
- Errors at point where handled (not logged and returned)
- Informational milestones during startup
- Debug info for tracing execution flow
- Successful operations with context for auditing

**Don't Log**:
- Errors that are returned (let caller decide)
- Sensitive data (passwords, tokens, full API keys)
- Noisy operations in tight loops

**Mask Sensitive Information**:
```go
// ✅ Good - masked
logger.Info("API key configured",
    logger.String("api_key_prefix", apiKey[:4]+"..."),
)

// ❌ Bad - exposed
logger.Info("API key", logger.String("api_key", apiKey))
```

## Package Organization

### Standard Layout

```
te-reo-bot/
├── cmd/                    # Entrypoints
│   ├── server/            # HTTP server
│   └── dict-gen/          # CLI tool
├── pkg/                    # Public libraries
│   ├── wotd/
│   ├── logger/
│   └── repository/
├── internal/               # Private code (if needed)
├── data/                   # Runtime data
└── specs/                  # Architecture docs
```

### Package Scope

**pkg/**: Reusable libraries, could be imported externally
**internal/**: Private code, cannot be imported externally
**cmd/**: Application entrypoints only

### Avoid Circular Dependencies

**Bad**:
```
pkg/wotd → pkg/repository → pkg/wotd  // Circular!
```

**Good**:
```
pkg/wotd → pkg/repository  // One-way dependency
```

Use interfaces to break circular dependencies.

## Interface Design

### Small, Focused Interfaces

**1-5 methods** per interface:
```go
// ✅ Good - focused
type WordRepository interface {
    GetAllWords() ([]Word, error)
    GetWordByID(id int) (*Word, error)
    AddWord(tx *sql.Tx, word *Word) error
}

// ❌ Bad - too many responsibilities
type DataManager interface {
    GetWords() ([]Word, error)
    SaveWord(word *Word) error
    ValidateWord(word *Word) error
    GenerateJSON() error
    UploadToGCS() error
    // ... 10 more methods
}
```

### Accept Interfaces, Return Structs

```go
// ✅ Good
func NewWordSelector(repo WordRepository) *WordSelector {
    return &WordSelector{repo: repo}
}

// ❌ Bad
func NewWordSelector(repo WordRepository) WordRepository {
    return &WordSelector{repo: repo}
}
```

### Document Interfaces

**Comprehensive godoc comments**:
```go
// WordRepository provides access to the word database.
// All methods are safe for concurrent use.
//
// GetWordByID retrieves a single word by its database ID.
// Returns sql.ErrNoRows if the word doesn't exist.
//
// AddWord inserts a new word within the provided transaction.
// The word's ID field will be populated with the generated ID.
type WordRepository interface {
    GetWordByID(id int) (*Word, error)
    AddWord(tx *sql.Tx, word *Word) error
}
```

## Git Workflow

### Signed Commits

**Always sign commits** with `-S` flag:
```bash
git commit -S -m "feat: add word validation"
```

### Conventional Commits

**Format**: `<type>: <description>`

**Types**:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `refactor`: Code refactoring (no behavior change)
- `test`: Adding or updating tests
- `chore`: Tooling, dependencies, build changes

**Examples**:
```
feat: add SQLite dictionary storage
fix: handle leap year day 366 correctly
docs: update README with CLI commands
refactor: extract validation logic to separate package
test: add integration tests for Mastodon client
```

### Branching

**Feature branches**:
```bash
git checkout -b feature/sqlite-migration
git checkout -b fix/leap-year-bug
```

**Worktrees** for parallel development:
```bash
git worktree add ../te-reo-bot-feature feature/sqlite-migration
cd ../te-reo-bot-feature
# Work on feature without context switching
```

## Extending Features

### Adding a New Social Media Platform

**Step-by-step**:

1. **Create client interface** in `pkg/wotd/`:
```go
// bluesky_client.go
type BlueskyClient struct {
    logger logger.Logger
    config BlueskyConfig
}

func (c *BlueskyClient) PostWord(word *WordOfTheDay) error {
    // Implementation
}
```

2. **Implement WordPoster interface**:
```go
type WordPoster interface {
    PostWord(word *WordOfTheDay) error
}
```

3. **Add configuration**:
```go
type BlueskyConfig struct {
    Server string `envconfig:"BLUESKY_SERVER"`
    Token  string `envconfig:"BLUESKY_TOKEN"`
}
```

4. **Update handler** in `pkg/handlers/messages_route.go`:
```go
case "bluesky":
    client := wotd.NewBlueskyClient(config)
    err = client.PostWord(word)
```

5. **Write tests**:
```go
// bluesky_client_test.go
func TestBlueskyClient_PostWord(t *testing.T) {
    // Table-driven tests
}
```

6. **Update documentation**:
- Add to README.md
- Update docs/constitution/product.md (product feature)
- Update docs/constitution/tech.md (technical details)

### Adding a New dict-gen Command

**Step-by-step**:

1. **Add command** to `cmd/dict-gen/main.go`:
```go
case "export-csv":
    exportCSV(flags)
```

2. **Implement handler**:
```go
func exportCSV(flags *Flags) {
    // Validate flags
    // Open database
    // Query words
    // Generate CSV
    // Write file
}
```

3. **Add to usage**:
```go
func printUsage() {
    fmt.Println("  export-csv   Export dictionary to CSV format")
}
```

4. **Write tests**:
```go
func TestExportCSV(t *testing.T) {
    // Test CSV generation
}
```

5. **Update documentation** in CLAUDE.md

## Reference Files

### Logging & Error Handling
- `pkg/logger/` - Custom logging implementation
- `pkg/entities/http-entities.go` - AppError and FriendlyError types
- `.kiro/specs/error-stack-traces/` - Stack trace system design

### Testing Examples
- `pkg/repository/repository_test.go` - Table-driven tests with testify
- `pkg/wotd/integration_test.go` - Integration test patterns
- `pkg/validator/validator_test.go` - Validation test examples

### Architecture
- `docs/constitution/tech.md` - Technical architecture details
- `specs/sqlite_dictionary_architecture/design.md` - SQLite migration design
