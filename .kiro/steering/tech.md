# Technology Stack & Build System

## Primary Technologies

- **Language**: Go 1.13+
- **Web Framework**: Gorilla Mux for HTTP routing
- **CLI Framework**: yacli for command-line interface
- **Testing**: testify/assert for unit tests
- **Configuration**: envconfig for environment-based config
- **Cloud Storage**: Google Cloud Storage SDK
- **Social Media APIs**: 
  - go-twitter for Twitter integration
  - go-mastodon for Mastodon integration

## Build System

The project uses **Make** as the primary build system with the following key targets:

### Common Commands

```bash
# Build static binary
make build-static

# Clean build artifacts
make clean

# Build Docker image
make docker-build

# Remove Docker images
make docker-rmi

# Update version
make version
```

### Running the Application

```bash
# Start the HTTP server
./te-reo-bot start-server -address="localhost" -port="8080" -tls="true"
```

## Development Tools

- **Node.js**: Used for JSON validation and formatting scripts
- **Docker**: Containerization support with multi-stage builds
- **Pre-commit hooks**: Code quality enforcement
- **GitHub Actions**: CI/CD pipeline

### JSON Validation Commands

```bash
# Validate dictionary structure
npm run validate-dictionary

# Lint JSON files
npm run validate-json

# Format JSON files
npm run format-json

# Test validation logic
npm run test-validation
```

## Dependencies

Key Go modules include gorilla/mux, kelseyhightower/envconfig, Google Cloud Storage client, and social media API clients. The project maintains minimal external dependencies for security and maintainability.