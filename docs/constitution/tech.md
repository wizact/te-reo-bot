# Te Reo Bot - Technical Constitution

## Technology Stack Rationale

### Go 1.13
**Why**: Existing codebase language with strong stdlib, excellent concurrency, single binary deployment.

**Pros**:
- Simple deployment (single static binary)
- Robust HTTP server support (net/http)
- Excellent tooling (go test, go build)
- Strong community for social media libraries

**Cons**:
- Older version (1.13), but stable for current needs
- Limited generic types support (pre-1.18)

**Alternatives Considered**: Python (rejected: heavier deployment), Node.js (rejected: less type safety)

### SQLite 3 (go-sqlite3 v1.14.24)
**Why**: Zero-config, file-based database suitable for ~366 word dataset.

**Pros**:
- No separate database server required
- ACID transactions for data integrity
- Simple backup (copy file)
- Sufficient performance for read-heavy workload

**Cons**:
- Single writer limitation (acceptable for CLI use)
- No network access (not needed)

**Alternatives Considered**: PostgreSQL (overkill), JSON files (completed, insufficient for word bank)

### Gorilla Mux v1.7.4
**Why**: Battle-tested HTTP routing with clean API.

**Pros**:
- Mature, stable library
- Route parameters and middleware support
- Well-documented

**Cons**:
- Larger than stdlib router
- Project in maintenance mode (acceptable for current use)

**Alternatives Considered**: stdlib http.ServeMux (too basic), chi (unnecessary change)

### Google Cloud Storage
**Why**: Reliable image hosting with public URL access.

**Pros**:
- Generous free tier
- CDN-backed delivery
- Simple API via official SDK
- Public URL support for social media

**Cons**:
- Cloud vendor lock-in
- Requires GCP credentials

**Alternatives Considered**: Local storage (rejected: not accessible to deployed server), S3 (similar cost/complexity)

### Social Media Libraries
**Twitter**: dghubble/go-twitter v0.0.0-20190719072343 + dghubble/oauth1 v0.6.0
**Mastodon**: mattn/go-mastodon v0.0.6

**Why**: Community-maintained clients with proven stability.

**Pros**:
- Handle OAuth complexity
- Type-safe API wrappers
- Active community support

**Cons**:
- Twitter library older (pre-API-v2)
- Mastodon library minimally maintained

### Custom Logging (pkg/logger)
**Why**: Zero external dependencies, full control over format and features.

**Pros**:
- Structured JSON logging
- Stack trace capture
- Request context tracking
- Environment-aware configuration
- No dependency bloat

**Cons**:
- Custom maintenance burden
- Reinvented wheel (many mature loggers exist)

**Rationale**: Project prioritizes minimal dependencies and specific security requirements (see Logging Standards below).

## Architecture Patterns

### Repository Pattern
**Implementation**: `pkg/repository/interface.go`

```go
type WordRepository interface {
    GetAllWords() ([]Word, error)
    GetWordsByDayIndex() (map[int]Word, error)
    AddWord(tx *sql.Tx, word *Word) error
    UpdateWord(word *Word) error
    DeleteWord(tx *sql.Tx, id int) error
    // ... transaction methods
}
```

**Rationale**:
- Abstract data access from business logic
- Enable testing with mock implementations
- Isolate SQLite specifics to one package

### Command Pattern (CLI)
**Implementation**: `cmd/dict-gen/main.go`

Subcommands: `migrate`, `generate`, `validate`

**Rationale**:
- Single binary with multiple operations
- Clear separation of concerns per command
- Easy to add new dictionary operations

### Generator Pattern
**Implementation**: `pkg/generator/generator.go`

Transforms SQLite records → dictionary.json

**Rationale**:
- Single responsibility (format transformation)
- Testable in isolation
- Enables future format variations

### Validator Pattern
**Implementation**: `pkg/validator/validator.go`

Ensures 366 unique words with complete data

**Rationale**:
- Data integrity before generation
- Clear error messages for corrections
- Prevents runtime failures from invalid data

### Middleware Chain (HTTP)
**Implementation**: `pkg/handlers/http-server.go`

1. Panic Recovery
2. API Key Authentication
3. Request Context Extraction
4. Business Logic Handler

**Rationale**:
- Separation of cross-cutting concerns
- Consistent error handling
- Request tracing and logging

## Detailed Architecture

### Data Flow

```
┌─────────────┐
│   SQLite    │ ← Source of truth (data/words.db)
│  words.db   │
└──────┬──────┘
       │
       ↓ dict-gen generate
┌─────────────────┐
│ dictionary.json │ ← Generated artifact
└──────┬──────────┘
       │
       ↓ Read on startup
┌─────────────────┐
│  HTTP Server    │
│  (in-memory)    │
└──────┬──────────┘
       │
       ↓ Calculate day, select word
┌─────────────────┐
│  WordSelector   │
└──────┬──────────┘
       │
       ├─→ Twitter API
       └─→ Mastodon API
```

### Component Interactions

**cmd/server**:
- Starts HTTP server via Gorilla Mux
- Initializes global logger
- Loads dictionary.json into memory
- Serves API endpoints: `/api/v1/messages`, `/healthcheck`

**cmd/dict-gen**:
- Standalone CLI tool
- Opens SQLite connection
- Executes subcommands (migrate/generate/validate)
- Minimal logging (fmt.Printf)
- No HTTP server

**pkg/wotd**:
- Business logic for word selection
- Social media client adapters
- Image acquisition from GCS
- Day-of-year calculation (1-366)

**pkg/repository**:
- SQLite connection management
- Transaction support
- CRUD operations
- Prepared statements (SQL injection prevention)

## Database Design

### Schema (pkg/repository/schema.go)

```sql
CREATE TABLE words (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    day_index INTEGER UNIQUE,  -- 1-366, nullable
    word TEXT NOT NULL,
    meaning TEXT NOT NULL,
    link TEXT,
    photo TEXT,
    photo_attribution TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT 1,

    CHECK (day_index IS NULL OR (day_index >= 1 AND day_index <= 366))
);

CREATE INDEX idx_day_index ON words(day_index);
CREATE INDEX idx_active ON words(is_active);
```

**Design Decisions**:
- **day_index nullable**: Allows word bank (words not yet assigned to days)
- **UNIQUE constraint**: Prevents duplicate day assignments
- **CHECK constraint**: Validates day range (1-366)
- **is_active flag**: Soft delete support (not hard deletes)
- **Indexes**: Fast lookups by day_index and active status

### Query Patterns

**Daily Word Selection** (HTTP server):
```go
// In-memory lookup from dictionary.json
word := wordMap[dayOfYear]
```

**Dictionary Generation** (dict-gen):
```sql
SELECT * FROM words
WHERE day_index IS NOT NULL AND is_active = 1
ORDER BY day_index
```

**Word Migration** (dict-gen):
```sql
BEGIN TRANSACTION;
UPDATE words SET day_index = NULL WHERE day_index IS NOT NULL;
-- For each word in dictionary.json:
SELECT * FROM words WHERE word = ?;
-- If exists: UPDATE, else: INSERT
COMMIT;
```

## Testing Requirements

### Framework
- **stdlib testing**: Built-in test runner
- **testify/assert**: Readable assertions
- **testify/require**: Critical assertions (stop on failure)

### Coverage Targets
- Unit tests: >80% coverage
- Integration tests for database operations
- End-to-end tests for HTTP handlers

### Test Organization
```
pkg/
  repository/
    repository_test.go          # Unit tests
    schema_test.go              # Schema validation
  wotd/
    word-selector_test.go       # Business logic
    integration_test.go         # Social media clients
  handlers/
    integration_test.go         # HTTP handlers
    panic_recovery_test.go      # Middleware
```

### Testing Patterns
- **Table-driven tests**: Multiple scenarios per test
- **In-memory SQLite**: Fast, isolated database tests (`:memory:`)
- **Test fixtures**: testdata/ directories for sample data
- **Mocks**: gomock for external service mocking

## Performance Targets

### HTTP Server
- **Request latency**: <10ms for word selection (in-memory lookup)
- **Startup time**: <1s (dictionary.json load)
- **Memory**: <50MB resident (small JSON file, no caching needed)

### dict-gen CLI
- **Generate command**: <500ms for 366 words
- **Validate command**: <100ms
- **Migrate command**: <1s for full dictionary import

### Database
- **SQLite file size**: <1MB for 1000 words
- **Query performance**: <1ms for indexed lookups
- **Transaction throughput**: Not critical (infrequent writes)

### Bottlenecks
- **Social media APIs**: Network latency (external, uncontrollable)
- **Image uploads**: GCS bandwidth (acceptable with CDN)

## Security & Operations

### Security Patterns

**API Authentication**:
- API key via `X-Api-Key` header
- Configured via environment variable
- Middleware validates before handler execution

**Error Handling**:
- Stack traces NEVER exposed in HTTP responses
- `AppError` type with internal context
- `FriendlyError` returned to clients
- Server-side logging with full details

**SQL Injection Prevention**:
- Prepared statements for all queries
- No string concatenation in SQL
- Parameterized queries via database/sql

**Secrets Management**:
- Environment variables for credentials
- No secrets in code or logs
- Masked sensitive data in structured logs

### Logging Standards

See [../conventions.md](../conventions.md) for complete logging and error handling patterns.

**Key Principles**:
1. Structured JSON logging (machine-parseable)
2. Request context tracking (request IDs)
3. Stack traces for debugging (server-side only)
4. Environment-aware configuration
5. Zero external logging dependencies

**Log Levels**:
- `debug`: Diagnostic details (dev environments)
- `info`: Operational events (default)
- `error`: Error conditions (with context)
- `fatal`: Unrecoverable errors (exit process)

### Deployment Architecture

**Google Cloud Run** (current deployment):
- Container-based deployment
- Automatic scaling (0-N instances)
- HTTPS with managed certificates
- Environment variable configuration

**Container Contents**:
- Go binary (te-reo-bot)
- dictionary.json (baked into image)
- Certificates for TLS (if self-hosted)

**Stateless Design**:
- No persistent storage in container
- dictionary.json read on startup (immutable)
- Images served from GCS (external)
- Horizontal scaling possible

**CI/CD** (GitHub Actions):
- Automated tests on PR
- Docker image build
- Deployment to Cloud Run
- Validation checks (dictionary integrity)

### Configuration Management

**Environment Variables**:
```bash
# Server
PORT=8080
ADDRESS=0.0.0.0
TLS_ENABLED=true
API_KEY=<secret>

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
ENABLE_STACK_TRACES=true
ENVIRONMENT=production

# GCS
GCS_BUCKET=te-reo-bot-images

# Social Media
TWITTER_API_KEY=<secret>
TWITTER_API_SECRET=<secret>
MASTODON_SERVER=https://mastodon.nz
MASTODON_TOKEN=<secret>
```

**Configuration Loading**:
- `kelseyhightower/envconfig` library
- Struct tags for environment mapping
- Validation on startup
- Fatal error if misconfigured

### Monitoring

**Current**:
- Structured JSON logs to stdout
- Cloud Run logging integration
- HTTP status codes for health checks

**Future Considerations**:
- Metrics collection (post success/failure rates)
- Alerting on posting failures
- Performance dashboards
- Error rate tracking

## Dependency Policy

### Minimal Dependencies
- Prefer stdlib where possible
- Evaluate each dependency for maintenance status
- Avoid transitive dependency bloat
- Pin versions in go.mod

### Current Dependencies
```
Direct:
- gorilla/mux v1.7.4          # HTTP routing
- mattn/go-sqlite3 v1.14.24   # SQLite driver (CGO)
- stretchr/testify v1.8.1     # Testing
- kelseyhightower/envconfig   # Config loading
- dghubble/go-twitter         # Twitter client
- mattn/go-mastodon v0.0.6    # Mastodon client
- cloud.google.com/go/storage # GCS client

Indirect:
- Standard library only for core logic
```

### Dependency Maintenance
- Monthly security updates
- Breaking change reviews before upgrading
- Test suite runs before dependency updates
- Document rationale for each dependency

### Future Considerations
- Migrate to Go 1.21+ (generics, improved stdlib)
- Replace Gorilla Mux with stdlib (Go 1.22+ router improvements)
- Evaluate Twitter API v2 migration
- Consider Mastodon library alternatives if maintenance stalls
