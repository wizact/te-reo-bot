# Te Reo Bot

M\u0101ori word-of-the-day social media bot. Posts daily M\u0101ori words with meanings and images to Mastodon and Twitter.

## Quick Reference

- **Product**: [docs/constitution/product.md](docs/constitution/product.md) - Vision, scope, roadmap
- **Technical**: [docs/constitution/tech.md](docs/constitution/tech.md) - Architecture, stack, patterns
- **Conventions**: [docs/conventions.md](docs/conventions.md) - Go coding standards
- **Features**: [docs/features/](docs/features/) - Feature planning templates

## Architecture

```
te-reo-bot/
├── cmd/
│   └── server/          # HTTP server (posts words via API)
├── pkg/
│   ├── wotd/           # Word-of-the-day logic
│   ├── repository/     # SQLite data access
│   ├── handlers/       # HTTP handlers
│   ├── logger/         # Structured logging
│   ├── storage/        # GCS image storage
│   └── entities/       # Domain models
├── data/
│   └── words.db        # SQLite database (source of truth)
└── docs/               # Architecture & feature documentation
```

**Data Flow**: SQLite (words.db) → HTTP server → Social media APIs

See [tech.md](docs/constitution/tech.md) for detailed architecture.

## Development Commands

```bash
# Build
make build

# Test
make test
make test-integration

# Run server
./te-reo-bot start-server -address="localhost" -port="8080"

# Database
# - Database schema auto-initializes on server startup
# - Production database: data/words.db (366 words)
# - Uses SQLite 3 with connection pooling
```

## Key Technologies

- **Go 1.13** - Primary language
- **SQLite 3** - Dictionary storage (via go-sqlite3)
- **Gorilla Mux** - HTTP routing
- **Google Cloud Storage** - Image hosting
- **Twitter/Mastodon APIs** - Social media posting
- **Custom Logger** - Structured JSON logging (pkg/logger)

See [tech.md](docs/constitution/tech.md) for technology decisions and rationale.

## Module Structure

**cmd/server**: HTTP API server
- Scheduled posts via GitHub Actions
- Reads from SQLite database (words.db)
- Posts to social media with images
- Auto-initializes database schema on startup

**pkg/repository**: Data access layer
- SQLite operations with Repository pattern
- Connection pooling for performance
- CRUD operations for words
- Returns domain models (wotd.Word)

**pkg/wotd**: Word-of-the-day business logic
- O(1) map-based word selection by day (1-366)
- O(1) word selection by index
- Social media client adapters (Twitter, Mastodon)
- Image acquisition and posting

See [tech.md](docs/constitution/tech.md) for detailed component architecture.
