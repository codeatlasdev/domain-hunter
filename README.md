<div align="center">

# domh

**Find your next domain in seconds, not hours.**

The fastest bulk domain availability checker. 1,437 TLDs. 19 registrars. Price comparison. One command.

[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev)
[![Release](https://img.shields.io/github/v/release/codeatlasdev/domain-hunter?style=flat-square&color=2563EB)](https://github.com/codeatlasdev/domain-hunter/releases)
[![License](https://img.shields.io/badge/license-MIT-green?style=flat-square)](LICENSE)
[![CI](https://img.shields.io/github/actions/workflow/status/codeatlasdev/domain-hunter/ci.yml?style=flat-square&label=CI)](https://github.com/codeatlasdev/domain-hunter/actions)

<img src="demo.gif" alt="domh demo" width="100%" />

[Install](#install) · [Quick Start](#quick-start) · [Features](#features) · [Presets](#presets) · [Docs](#usage)

</div>

---

## Why domh?

| | domh | Others |
|---|---|---|
| **Speed** | ~500 domains/sec | ~1-10/sec |
| **TLDs** | 1,437 (auto-updated from IANA) | 5-50 hardcoded |
| **Price comparison** | 19 registrars + buy links | ❌ |
| **Verification** | DNS + RDAP + WHOIS + SSL | Usually just WHOIS |
| **TUI** | Split-panel dashboard | Plain text |
| **Self-update** | `domh update` | Re-download manually |

---

## Install

```bash
# Go (recommended)
go install github.com/codeatlasdev/domain-hunter@latest

# macOS / Linux
curl -fsSL https://raw.githubusercontent.com/codeatlasdev/domain-hunter/main/install.sh | sh

# Or download from releases
# → https://github.com/codeatlasdev/domain-hunter/releases
```

---

## Quick Start

```bash
# Interactive wizard — just run it
domh

# Find 4-letter .com domains
domh scan --tld com --length 4

# Check a brand across startup TLDs
domh scan coolname --preset startup

# Dictionary mode — check your name ideas
echo -e "nexo\nhivo\nzuno\ntevo" > names.txt
domh check names.txt --tld com,dev,io

# Prefix/suffix combos
domh scan myapp --prefix get,try --suffix hub,ly --tld com,dev

# Preview without checking
domh scan mybrand --preset tech --dry-run

# CI pipeline (no TUI, JSON output)
domh scan --length 4 --tld com --batch --json --yes > results.json
```

---

## Features

### ⚡ Blazing Fast
DNS-first strategy with 16 public resolvers (Cloudflare, Google, Quad9, OpenDNS, etc). RDAP confirmation only for candidates. ~500 domains/second with 50 workers.

### 🌍 Every TLD on Earth
1,437 TLDs auto-fetched from IANA. 1,199 with RDAP support. Auto-refreshes every 24h. Check `--all` to scan every single one.

### 💰 Price Comparison
Instantly shows cheapest registrar + price for available domains. 19 registrars with direct buy links:

> Namecheap · Porkbun · Cloudflare · GoDaddy · Google · Dynadot · NameSilo · Spaceship · Hostinger · IONOS · OVH · Gandi · Hover · Epik · Name.com · Registro.br · 101domain · INWX · Dreamhost

### 🔍 4-Layer Verification
1. **DNS** (NS → A → MX) — fast, no rate limit
2. **RDAP** — authoritative confirmation
3. **WHOIS** — multi-server fallback
4. **SSL** — last resort for edge cases

### 🎨 Beautiful TUI
Split-panel dashboard that uses 100% of your terminal. Responsive. Shows progress, stats, available domains with prices, and activity log — all at once.

### 🧠 Smart Generation
Pronounceable patterns only — no garbage like `xqz.com`:

| Pattern | Example | Count |
|---------|---------|-------|
| CVC | bun, dev, kit | 1,805 |
| VCV | ava, elo, umi | 475 |
| CVCV | buno, tevo, kaze | 9,025 |
| CVCVC | nexus, pixel | 171,475 |

### 📦 Presets

```bash
domh presets  # list all

# Built-in:
startup    → com, org, io, ai, tech, app, dev, xyz
tech       → io, ai, app, dev, tech, cloud, software, code, systems
creative   → design, art, studio, media, photography, gallery, ink
ecommerce  → shop, store, market, sale, deals, buy, shopping
finance    → finance, capital, fund, money, investments, bank, pay
country    → us, uk, de, fr, ca, au, br, in, nl, jp
brazil     → com.br, net.br, org.br, app.br, dev.br
# + 5 more
```

---

## Usage

```
◆ domh — bulk domain availability checker

Usage:
  domh                       Interactive wizard
  domh scan [name] [flags]   Scan domains
  domh check <file> [flags]  Dictionary mode
  domh tlds [flags]          List TLDs
  domh presets               List presets
  domh update                Self-update
  domh version               Version info

Scan flags:
  --tld          TLDs (comma-separated)       [default: com]
  --preset       Use preset TLD set
  --all          Check ALL 1,437 TLDs
  --length       Domain length (3-5)          [default: 3]
  --pattern      CVC, VCV, CVCV, ALL          [default: ALL]
  --prefix       Prefixes (comma-separated)
  --suffix       Suffixes (comma-separated)
  --workers      Concurrent workers           [default: 50]
  --format       Export: txt,json,csv         [default: txt]
  --regex, -r    Regex filter
  --delay        Delay between queries (ms)   [default: 0]
  --info         Show registrar info for taken domains
  --dry-run      Preview without checking
  --batch        Plain output (no TUI)
  --yes, -y      Skip confirmations
  --force        Skip performance warnings
```

---

## How It Works

<div align="center">
<img src="assets/how-it-works.svg" alt="How domh works" width="100%" />
</div>

---

## Output

Results are saved in real-time (won't lose data if you quit early):

- **txt** — one domain per line
- **json** — includes pricing, registrar, buy URLs
- **csv** — `domain,tld,available,cheapest_price,cheapest_registrar,buy_url`

---

## Self-Update

```bash
domh update
# ✓ Updated domh v1.4.0 → v1.5.0
```

---

## vs. Alternatives

| Feature | domh | domain-check (Rust) | domain-scanner (Go) | domaindex.io |
|---------|------|--------------------|--------------------|-------------|
| Speed | ~500/s | ~100/s | ~1/s | ~30/s |
| TLDs | 1,437 | 1,200 | 1 | 100+ |
| Price comparison | ✅ 19 registrars | ❌ | ❌ | ❌ |
| Buy links | ✅ | ❌ | ❌ | ✅ |
| TUI dashboard | ✅ split-panel | ❌ spinner | ❌ | web only |
| Interactive wizard | ✅ | ❌ | ❌ | ❌ |
| Self-update | ✅ | ❌ | ❌ | N/A |
| DNS verification | ✅ NS+A+MX | ❌ | ✅ | ❌ |
| WHOIS fallback | ✅ multi-server | ✅ | ✅ | ❌ |
| SSL check | ✅ | ❌ | ✅ | ❌ |
| Presets | ✅ 12 | ✅ 11 | ❌ | ❌ |
| Prefix/suffix | ✅ | ✅ | ❌ | ❌ |
| Regex filter | ✅ | ❌ | ✅ | ❌ |
| CI/batch mode | ✅ | ✅ | ❌ | ❌ |
| Offline fallback | ✅ 32 TLDs | ✅ | ❌ | ❌ |
| MCP server | 🔜 | ✅ | ❌ | ❌ |

---

## Contributing

PRs welcome. Run `go build ./...` and `go test ./...` before submitting.

## License

MIT — [CodeAtlas](https://codeatlas.com.br)
