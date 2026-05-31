FROM golang:1.17-alpine AS build

# Install build dependencies including gcc for CGO (required by go-sqlite3)
RUN apk add --no-cache \
	bash \
	git \
	make \
	gcc \
	musl-dev

WORKDIR /build

COPY . .

# Enable CGO for SQLite driver
ENV CGO_ENABLED=1

RUN make build-static




FROM alpine:3.11 AS base

LABEL maintainer="amir.mohtasebi@gmail.com"

# Install runtime dependencies for SQLite
RUN set -x \
    && apk --update add ca-certificates wget gnupg sqlite-libs && rm -rf /var/cache/apk/

EXPOSE 8080

WORKDIR /app/

# Copy binary from build stage
COPY --from=build /build/out/ .

# Copy SQLite database (source of truth)
COPY --from=build /build/data/words.db ./data/

# Create data directory with proper permissions
RUN mkdir -p data && chmod 755 data

# Set database path for container environment
ENV DB_PATH=/app/data/words.db

ENTRYPOINT ["./te-reo-bot"]
CMD ["start-server", "-address=localhost", "-port=8080", "-tls=false"]
