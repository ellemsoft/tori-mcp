# tori

Search [Tori.fi](https://www.tori.fi) (Finnish marketplace) from the command line.
Stdlib + mcp-go for MCP protocol.

## Install

```bash
go build -o tori .
mv tori /usr/local/bin/  # optional
```

## Usage

```bash
# Basic search
tori search iphone

# With filters
tori search iphone --price-from 100 --price-to 500 --condition good
tori search sähköpyörä --filter bikes_type=8 --filter condition=2

# Category + location discovery
tori categories puhelin
tori categories 1.69.3963   # drill down into Cycling
tori locations helsinki

# See what filters are available
tori filters iphone

# Listing details
tori show 44583092
tori show 44583092 --json
tori show 44583092 --fetch-body --json  # includes description + detail tags
```

## Commands

| Command | Description |
|---------|-------------|
| `search <query>` | Search listings (table: ID, type, price, location, title, URL) |
| `filters <query>` | Show available filter options for a query |
| `categories [query\|code]` | Browse category tree; pass code to see children |
| `locations [query]` | Browse location tree |
| `show <id>` | Listing details (--json for structured, --fetch-body for description) |

## Flags

### search
| Flag | Description |
|------|-------------|
| `--category CODE` | Sub-category code |
| `--location CODE` | Location code |
| `--price-from N` | Min price EUR |
| `--price-to N` | Max price EUR |
| `--condition TYPE` | new \| like-new \| good \| fair |
| `--shipping` | ToriDiili items only |
| `--filter K=V` | Raw filter param (repeatable) |
| `--page N` | Page number (default 1) |
| `--lang LANG` | en (default) \| fi |
| `--json` | Structured JSON output |
| `--raw` | Raw API JSON passthrough |

### show
| Flag | Description |
|------|-------------|
| `--json` | Structured JSON output |
| `--fetch-body` | Fetch description + detail tags from listing page |

## Output formats

**Table** (default): compact, with URLs, total count.

**JSON** (`--json`): typed struct with all API fields.

**Raw** (`--raw`): direct API JSON — max flexibility for LLM consumption.

**Show with body** (`--fetch-body --json`): includes `description` field and `details` map (condition, size, type, etc.) scraped from the listing page. Adds ~500ms.

## Architecture

```
main.go          CLI routing, --help, flag parsing
client.go        HTTP client with iOS headers + finn-gw-key signing
signing.go       HMAC-SHA512 key computation
search.go        Search, filters, categories, locations, show, page scraping
display.go       Table + JSON output, bilingual labels
categories_en.go 120+ category translations (fi→en)
```

## No auth required

## Credits

API endpoints and signing key discovered by [torium](https://github.com/ahnl/torium) — the original Tori.fi Python client and MCP server.

Search is public. Uses the same endpoints as the Tori iOS app.

## MCP Server

The CLI doubles as an MCP server:

```bash
tori --mcp :8081          # start MCP server
tori --mcp 127.0.0.1:8081 # bind to specific address
```

Tools: `search`, `show`, `filters`, `categories`, `locations`

Rate limited: 120 req/min per IP. Stats persisted to `/opt/tori/stats.json`.

### Deploy

```bash
# Cross-compile + upload
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o tori .
scp tori root@<host>:/opt/tori/

# systemd
systemctl enable --now tori-mcp
```

Systemd unit + Caddy config in `deploy/`. Full guide in `deploy/DEPLOY.md`.

## Disclaimer

This is an independent, community-developed search client. It is not affiliated with, endorsed by, or sponsored by Tori.fi, Vend Marketplaces Oy, or Schibsted. Use is subject to the terms of the respective source services. No listing data is stored.

MIT licensed — see [LICENSE](LICENSE).
