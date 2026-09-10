---
name: domain-hunter
description: Check domain availability, generate name candidates, compare prices across 20+ registrars, check Brazilian (.br) domains via Registro.br, get trademark search URLs (INPI/USPTO/EUIPO). Use for startup naming, SaaS branding, bulk domain scanning, agent-driven naming workflows, and Brazilian market research.
---

# Domain Hunter

Bulk domain availability checker with Brazilian market support, price comparison, and AI agent-friendly output.

## When to use

- Find available domains for a SaaS, startup, or brand
- Check a shortlist of name candidates across multiple TLDs at once
- Generate pronounceable name candidates (3-5 chars) then check availability
- Compare registration prices across registrars (including Registro.br for .br TLDs)
- Check Brazilian market: .com.br, .app.br, .dev.br, and all other .br SLDs
- Get trademark search links (INPI, USPTO, EUIPO) for candidate names
- CI/CD pipelines or agent workflows that need machine-readable JSON output

## Trigger phrases

"find a domain", "is this domain available", "check domain", "domain name for my SaaS",
"short domain", "4 letter domain", "pronounceable name", "startup name", "brand name",
"domínio disponível", ".com.br available", "registrar domínio", "verificar domínio",
"check .br domain", "brazil domain", "INPI trademark", "marca registrada",
"domain price", "cheapest registrar", "where to register", "bulk domain check"

## Setup

### MCP (recommended for agent use)

```bash
go install github.com/codeatlasdev/domain-hunter@latest
```

Add to your MCP config (`~/.config/omp/mcp.json` or equivalent):

```json
{
  "mcpServers": {
    "domh": {
      "command": "domh",
      "args": ["mcp"]
    }
  }
}
```

### CLI (for direct use or piping)

```bash
go install github.com/codeatlasdev/domain-hunter@latest
```

## MCP Tools

### scan_names — primary tool for AI naming workflows

Check one or more base names across multiple TLDs in one call. Returns structured JSON with availability, pricing, and trademark search URLs.

```
names    comma-separated base names (e.g. "kora,nexus,velo,plex")
tlds     comma-separated TLDs (e.g. "com,io,app,com.br") — optional if using preset
preset   TLD preset name — optional if using tlds
trademark  include INPI/USPTO/EUIPO search URLs (default: true)
```

Presets: `saas`, `startup`, `brazil`, `brazil-full`, `brazil-saas`, `global-saas`, `tech`, `popular`, `classic`, `enterprise`, `creative`, `ecommerce`, `finance`, `web`, `trendy`, `country`, `br-pro`

Response schema:
```json
{
  "query": { "names": [...], "tlds": [...] },
  "available": [
    {
      "domain": "kora.io",
      "name": "kora",
      "tld": "io",
      "available": true,
      "method": "dns+rdap",
      "pricing": { "cheapest": { "registrar": "Porkbun", "price": 28.88, "currency": "USD", "buy_url": "..." } },
      "trademark": { "inpi": "...", "uspto": "...", "euipo": "..." }
    }
  ],
  "taken": [...],
  "errors": [...],
  "stats": { "checked": 12, "available": 3, "taken": 8, "errors": 1 }
}
```

### check_domain

Check a single full domain name.

```
domain  full domain (e.g. "coolname.com", "myapp.com.br")
```

### check_domains

Check multiple full domains at once.

```
domains  comma-separated (e.g. "cool.com,cool.dev,cool.com.br")
```

### check_with_preset

Check one base name across a curated TLD set.

```
name    base name (e.g. "kora")
preset  preset name (e.g. "global-saas", "brazil-full")
```

### generate_names

Generate pronounceable name candidates by length and phonetic pattern. Use this before `scan_names` when you need candidates.

```
length   3, 4, or 5
pattern  CVC | VCV | CVCV | CVCVC | ALL (default: ALL)
tld      TLD to append for preview (default: com)
```

Returns up to 100 domain candidates.

### get_prices

Get registrar prices and buy links for a domain.

```
domain  full domain (e.g. "coolname.com", "myapp.com.br")
```

Returns prices from 20+ registrars sorted cheapest first. .br TLDs return Registro.br price in BRL (R$40/year).

### trademark_urls

Get trademark search links for a name across INPI (Brazil), USPTO (US), and EUIPO (EU).

```
name  base name (e.g. "kora")
```

### list_presets

List all available TLD presets and their TLD lists.

## CLI — agent workflow

### suggest command (JSON output, agent-ready)

Takes name candidates, checks across TLDs, returns structured JSON:

```bash
domh suggest kora nexus velo plex --preset global-saas
```

```bash
echo -e "kora\nnexus\nvelo\nplex" | domh suggest --preset brazil-full
```

```bash
domh suggest kora --tld com,io,app,com.br,app.br --stream   # NDJSON, one result per line
```

### scan command (bulk pattern generation)

```bash
domh scan --length 4 --preset saas --batch --json           # NDJSON to stdout
domh scan myapp --tld com,io,app,com.br --batch             # check one name, all TLDs
domh scan --length 5 --pattern CVCVC --preset brazil --batch
```

### check command (dictionary file)

```bash
domh check names.txt --tld com,io,app,com.br --batch
```

## Typical agent workflow

1. Generate candidates with `generate_names` (pronounceable, 4-5 chars)
2. Check availability with `scan_names` + preset `global-saas` or `brazil-full`
3. For available names, call `trademark_urls` to surface INPI/USPTO/EUIPO links
4. Call `get_prices` on the best candidate for registrar comparison

Or in one step: `scan_names` with `trademark: true` (default) returns everything at once.

## Brazilian market notes

- All .br SLDs (com.br, net.br, app.br, dev.br, etc.) are checked via `rdap.registro.br` — no extra setup needed
- Only NIC.br (Registro.br) registers .br domains — R$40/year flat, no third-party registrars
- INPI trademark search is link-based (no public API exists); `trademark_urls` returns the direct search URL
- Presets: `brazil` (5 TLDs), `brazil-full` (15 TLDs), `brazil-saas` (4 TLDs), `br-pro` (professional TLDs)
