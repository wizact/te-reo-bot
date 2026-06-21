# Te Reo Twitter Bot

![Go Report](https://goreportcard.com/badge/github.com/wizact/te-reo-bot) ![Go](https://github.com/wizact/te-reo-bot/workflows/Go/badge.svg)

māori -te reo- word of the day mastodon bot.

## Documentation

### For Developers

- **[CLAUDE.md](CLAUDE.md)** - Quick reference and navigation hub
- **[docs/constitution/product.md](docs/constitution/product.md)** - Product vision, scope, and roadmap
- **[docs/constitution/tech.md](docs/constitution/tech.md)** - Technical architecture and decisions
- **[docs/conventions.md](docs/conventions.md)** - Go coding standards and patterns
- **[docs/features/](docs/features/)** - Feature specification templates

**Start here**: New contributors should read [CLAUDE.md](CLAUDE.md) first for quick orientation.

## Usage

```bash
./te-reo-bot start-server -address="localhost" -port="8080" -tls="true"
```

## Curator TUI

Build the local curator tool:

```bash
make build-curator
```

Run the keyboard-first curator UI against the default database:

```bash
./out/te-reo-curator
```

Run validation only:

```bash
./out/te-reo-curator -validate
```

### Curator Shortcuts

- `↑` / `↓` or `j` / `k` - move selection
- `/` - filter by text
- `c` - clear filter
- `s` - cycle sort column
- `g` - toggle ascending/descending sort
- `a` - add a word
- `e` - edit selected word metadata
- `d` - assign a day index
- `u` - clear selected day index
- `n` - auto-assign next available day
- `v` - run validation
- `r` - reload from SQLite
- `?` - show shortcut help
- `q` - quit
