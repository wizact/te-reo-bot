# Docker Guide

## Building the Image

```bash
# Build the image
docker build -t te-reo-bot:latest .

# Or using make
make docker-build
```

## Running the Container

### Basic Usage

```bash
docker run -p 8080:8080 te-reo-bot:latest
```

### With Environment Variables

```bash
docker run -p 8080:8080 \
  -e TEREOBOT_APIKEY="your-api-key" \
  -e TEREOBOT_BUCKETNAME="your-gcs-bucket" \
  -e TEREOBOT_DRYRUN=false \
  te-reo-bot:latest
```

### With Custom Database

Mount your own database file:

```bash
docker run -p 8080:8080 \
  -v /path/to/your/words.db:/app/data/words.db:ro \
  -e DB_PATH=/app/data/words.db \
  te-reo-bot:latest
```

### With Persistent Data

To persist database changes (if using writable mode):

```bash
docker run -p 8080:8080 \
  -v te-reo-data:/app/data \
  te-reo-bot:latest
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_PATH` | Path to SQLite database | `/app/data/words.db` |
| `TEREOBOT_APIKEY` | API key for authentication | Required |
| `TEREOBOT_BUCKETNAME` | GCS bucket name | Required |
| `TEREOBOT_DRYRUN` | Dry-run mode (true/false) | `false` |
| `TEREOBOT_CONSUMERKEY` | Twitter consumer key | Optional |
| `TEREOBOT_CONSUMERSECRET` | Twitter consumer secret | Optional |
| `TEREOBOT_ACCESSTOKEN` | Twitter access token | Optional |
| `TEREOBOT_ACCESSSECRET` | Twitter access secret | Optional |
| `TEREOBOT_MASTODONSERVERNAME` | Mastodon server URL | Optional |
| `TEREOBOT_MASTODONCLIENTID` | Mastodon client ID | Optional |
| `TEREOBOT_MASTODONACCESSTOKEN` | Mastodon access token | Optional |

## Docker Compose Example

```yaml
version: '3.8'

services:
  te-reo-bot:
    image: te-reo-bot:latest
    ports:
      - "8080:8080"
    environment:
      - TEREOBOT_APIKEY=${API_KEY}
      - TEREOBOT_BUCKETNAME=${GCS_BUCKET}
      - TEREOBOT_DRYRUN=false
      - DB_PATH=/app/data/words.db
    volumes:
      - ./data/words.db:/app/data/words.db:ro
    restart: unless-stopped
```

## Build Requirements

The Dockerfile uses multi-stage builds:

### Build Stage
- **Base**: `golang:1.17-alpine`
- **Dependencies**: 
  - `gcc` and `musl-dev` for CGO (required by go-sqlite3)
  - `make`, `git`, `bash` for build process
- **CGO**: Enabled for SQLite support

### Runtime Stage
- **Base**: `alpine:3.11`
- **Dependencies**:
  - `ca-certificates` for HTTPS
  - `sqlite-libs` for SQLite runtime
  - `wget`, `gnupg` for utilities

## Database Management

### Viewing Database in Container

```bash
# Shell into running container
docker exec -it <container-id> sh

# Query database
apk add sqlite
sqlite3 /app/data/words.db "SELECT COUNT(*) FROM words;"
```

### Backing Up Database

```bash
# Copy database from container
docker cp <container-id>:/app/data/words.db ./backup-words.db
```

### Restoring Database

```bash
# Stop container
docker stop <container-id>

# Run with new database
docker run -p 8080:8080 \
  -v /path/to/backup-words.db:/app/data/words.db:ro \
  te-reo-bot:latest
```

## Troubleshooting

### Database Not Found

**Error**: `Failed to open database connection`

**Solution**: Ensure `DB_PATH` environment variable points to correct location or mount database at `/app/data/words.db`.

### Permission Denied

**Error**: `unable to open database file: unable to open database file`

**Solution**: Check file permissions on mounted database:
```bash
chmod 644 /path/to/words.db
```

### Binary Not Running

**Error**: `standard_init_linux.go: exec user process caused: no such file or directory`

**Solution**: Ensure CGO is enabled during build. Rebuild with:
```bash
docker build --no-cache -t te-reo-bot:latest .
```

## Image Size Optimization

The multi-stage build keeps the final image small:
- Build stage: ~500MB (discarded)
- Final image: ~20MB

**Tips:**
- Use `.dockerignore` to exclude unnecessary files
- Database is ~72KB, included in image
- Only production database copied, test files excluded
