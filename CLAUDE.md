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
│   ├── server/          # HTTP server (posts words via API)
│   └── dict-gen/        # CLI tool (manage SQLite dictionary)
├── pkg/
│   ├── wotd/           # Word-of-the-day logic
│   ├── repository/     # SQLite data access
│   ├── generator/      # JSON generation
│   ├── validator/      # Data validation
│   ├── migration/      # Dict import/export
│   ├── handlers/       # HTTP handlers
│   ├── logger/         # Structured logging
│   ├── storage/        # GCS image storage
│   └── entities/       # Domain models
├── data/
│   └── words.db        # SQLite database (source of truth)
└── specs/              # Architecture documents
```

**Data Flow**: SQLite → dict-gen generate → dictionary.json → HTTP server → Social media APIs

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

# Manage dictionary
./dict-gen migrate --input=dictionary.json    # Import to SQLite
./dict-gen validate                            # Check integrity (366 words)
./dict-gen generate --output=dictionary.json  # Export from SQLite
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
- Reads dictionary.json (generated artifact)
- Posts to social media with images

**cmd/dict-gen**: Dictionary management CLI
- Import/export dictionary.json ↔ SQLite
- Validate data integrity (366 unique words)
- Manage word lifecycle

**pkg/repository**: Data access layer
- SQLite operations with Repository pattern
- Transaction support for migrations
- CRUD operations for words

**pkg/wotd**: Word-of-the-day business logic
- Day-based word selection (1-366)
- Social media client adapters
- Image acquisition and posting

See [tech.md](docs/constitution/tech.md) for detailed component architecture.
