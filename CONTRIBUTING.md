# Contributing

PRs welcome. Here's how:

```bash
git clone https://github.com/codeatlasdev/domain-hunter
cd domain-hunter
go build ./...
go test ./...
```

## Before submitting

```bash
go build ./...
go test ./...
go vet ./...
```

## Structure

- `main.go` — CLI entry point
- `cmd/mcp/` — MCP server binary
- `internal/scanner/` — domain checking engine
- `internal/pricing/` — registrar prices
- `internal/presets/` — TLD presets
- `internal/registry/` — IANA TLD + RDAP bootstrap
- `internal/tui/` — bubbletea TUI
- `internal/export/` — result export (txt/json/csv)
- `web/api/` — HTTP API server
- `web/app/` — React frontend

## Adding a new registrar

Edit `internal/pricing/pricing.go`:
1. Add buy link function to `buyLinks`
2. Add prices to `knownPrices` for supported TLDs

## Adding a new preset

Edit `internal/presets/presets.go` — add entry to the `Presets` map.
