# AGENTS.md — tori-cli

Search Tori.fi from the command line. One binary, one dependency (mcp-go for MCP protocol).

## Build

```bash
go build -o tori .
```

## Architecture

```
main.go      CLI entry, subcommand routing, flag parsing
client.go    HTTP client with iOS app headers + finn-gw-key signing
signing.go   HMAC-SHA512 key computation for API gateway
search.go    Search, categories, locations, filters, show — all data types + API calls
display.go   Output formatting (tables, JSON)
```

## Conventions

- **Stdlib + mcp-go.** No external modules. `net/http`, `encoding/json`, `text/tabwriter`, `flag`, `crypto/*`.
- **LLM-friendly.** The core use case is an LLM invoking this CLI. Default table output is compact and parseable. `--json` for structured data, `--raw` for passthrough API JSON.
- **No auth.** Public search endpoints work with only the signing key + iOS app headers. No bearer tokens needed.
- **One round-trip per command.** Search, categories, and locations each make exactly one API call (search also fetches pole position in parallel).
- **Types over `any`.** All API responses are parsed into typed structs. `map[string]any` is only used at the HTTP boundary.
- **Go 1.22+ idioms.** `for i := range n` for counting loops.

## API notes

- Base URL: `https://apps-gw-poc.svc.tori.fi`
- All requests need `finn-gw-key` (HMAC-SHA512, key = UUID `3b535f36-...`) and iOS app headers.
- Search key: `SEARCH_ID_BAP_COMMON`, service: `SEARCH-QUEST-RC`.
- Price uses `currency_code` field (e.g. `"EUR"`), not `currency`.
- Filter hit counts use `hits` field, not `count`.
- Ad type (Myydään/Ostetaan/Annetaan) inferred from heading prefix since the API doesn't distinguish in search results.
- Promoted listings come from a separate pole-position endpoint.

## Testing

```bash
go build -o tori . && ./tori search iphone
./tori search iphone --json
./tori search iphone --raw
./tori search iphone --price-from 100 --price-to 500 --shipping
./tori categories puhelin
./tori locations helsinki
./tori filters iphone
./tori show 44583092
```

## Testing

```bash
# Fast: unit + mock + fixture tests (~1s, no network)
go test ./... -count=1

# Slow: end-to-end against live API (~75s)
bash testcases/run_e2e.sh
```


**Mock tests** (`client_test.go`) use `net/http/httptest` with hand-crafted responses.

**Unit tests** (`search_test.go`) test pure logic: signing, query encoding, parsing, labels.
