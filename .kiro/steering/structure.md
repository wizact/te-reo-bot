# Project Structure & Organization

## Directory Layout

The project follows Go standard project layout conventions:

```
├── cmd/                    # Main applications
│   └── server/            # HTTP server application
│       ├── main.go        # Application entry point
│       ├── start-server-command.go  # CLI command implementation
│       └── dictionary.json # Māori word dictionary data
├── pkg/                   # Library code
│   ├── entities/          # Data structures and models
│   ├── handlers/          # HTTP request handlers
│   ├── storage/           # Cloud storage integration
│   └── wotd/             # Word-of-the-day logic
├── scripts/              # Build and validation scripts
├── http-client/          # HTTP client test files
├── certs/                # TLS certificates
├── out/                  # Build output directory
└── version/              # Version information
```

## Code Organization Patterns

### Package Structure
- **cmd/**: Contains main applications, each in its own subdirectory
- **pkg/**: Reusable library code organized by domain
- **Internal packages**: Use descriptive names (handlers, entities, storage, wotd)

### Naming Conventions
- **Files**: Use kebab-case for multi-word files (`http-server.go`, `word-selector.go`)
- **Packages**: Use lowercase, single words when possible
- **Tests**: Follow `*_test.go` pattern with `package_test` naming

### Configuration
- Environment-based configuration using `envconfig` tags
- TLS certificates stored in `certs/` directory
- Dictionary data co-located with server application in `cmd/server/`

## Key Files
- **dictionary.json**: Core data file containing Māori words and metadata
- **Makefile**: Build automation and common tasks
- **Dockerfile**: Multi-stage container build
- **VERSION.txt**: Version tracking for releases

## Testing Structure
- Unit tests alongside source code (`*_test.go`)
- Validation scripts in `scripts/` directory
- HTTP client tests in `http-client/` directory
- Use testify/assert for test assertions