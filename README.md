# ◆ Domain Hunter

Bulk domain availability checker powered by RDAP. Beautiful TUI, blazing fast.

![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/license-MIT-blue)
![Release](https://img.shields.io/github/v/release/codeatlasdev/domain-hunter?color=blue)

```
╭──────────────────────────────────────────────────╮
│  ◆ Domain Hunter v2                              │
│    4-letter .com,.dev • 80 workers • RDAP        │
╰──────────────────────────────────────────────────╯

╭──────────────────────────────────────────────────────────────╮
│  ⠋ 1842/32400  ━━━━━━━━━━━━━━━━━━━━────────────────────────  │
│                                                              │
│  available 23    taken 1814    errors 5    elapsed 18s       │
│  eta 5m12s       rate 102/s                                  │
╰──────────────────────────────────────────────────────────────╯

╭──────────────────────────────────────────────────╮
│  ● Available                                     │
│                                                  │
│    ✓ buxo.com                                    │
│    ✓ kevi.dev                                    │
│    ✓ zupo.com                                    │
╰──────────────────────────────────────────────────╯
```

## Install

```bash
# macOS / Linux (Homebrew)
brew install codeatlasdev/tap/domain-hunter

# Go install
go install github.com/codeatlasdev/domain-hunter@latest

# Download binary
curl -fsSL https://raw.githubusercontent.com/codeatlasdev/domain-hunter/main/install.sh | sh
```

## Usage

### Interactive Mode (wizard)

Just run without arguments:

```bash
domain-hunter
```

You'll get a beautiful wizard to configure:
1. **TLDs** — multi-select (.com, .dev, .io, .app, .net, .org, .co, .xyz)
2. **Length** — 3, 4, or 5 letters
3. **Pattern** — CVC, VCV, CVCV, CVCVC, or ALL
4. **Workers** — concurrent goroutines (default: 50)
5. **Export** — txt, json, csv

### CLI Mode (direct)

```bash
# 4-letter .com domains, CVCV pattern, 80 workers
domain-hunter scan --tld com --length 4 --pattern cvcv --workers 80

# Multiple TLDs, export as JSON
domain-hunter scan --tld com,dev,io --length 3 --format json

# 5-letter domains, all patterns
domain-hunter scan --tld com --length 5 --pattern all --workers 100
```

## How It Works

Uses **RDAP** (Registration Data Access Protocol) — the modern replacement for WHOIS:
- HTTP GET to the TLD's RDAP server
- `404` = domain available
- `200` = domain taken
- No parsing needed, no rate limit issues
- ~100+ domains/second with 50 workers

### Supported TLDs

| TLD | Provider |
|-----|----------|
| .com | Verisign RDAP |
| .net | Verisign RDAP |
| .org | PIR RDAP |
| .dev | Google RDAP |
| .app | Google RDAP |
| .io | NIC.io RDAP |
| .co | NIC.co RDAP |
| .xyz | NIC.xyz RDAP |

## Domain Patterns

| Pattern | Example | Count (per TLD) |
|---------|---------|-----------------|
| CVC | bun, dev, kit | ~1,900 |
| VCV | ava, elo, umi | ~475 |
| CVCV | buno, tevo | ~36,100 |
| CVCVC | nexus, pixel | ~684,000 |
| ALL | all above | varies |

## Export Formats

Results are saved in real-time (won't lose data if you quit early):

- **txt** — one domain per line
- **json** — `[{"domain": "bun.com", "tld": "com", "checked_at": "..."}]`
- **csv** — `domain,tld,checked_at`

Files: `results-{timestamp}.{ext}`

## Keybindings

| Key | Action |
|-----|--------|
| `q` / `Ctrl+C` | Quit (saves partial results) |
| `Esc` | Quit |

## Development

```bash
git clone https://github.com/codeatlasdev/domain-hunter
cd domain-hunter
go build -o domain-hunter .
./domain-hunter
```

## Release

Tags trigger automatic multi-platform builds via GoReleaser:

```bash
git tag v1.0.0
git push --tags
```

Builds for: Linux (amd64, arm64), macOS (amd64, arm64), Windows (amd64, arm64).

## License

MIT — [CodeAtlas](https://codeatlas.com.br)
