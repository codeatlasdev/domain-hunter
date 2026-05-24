---
name: domain-hunter
description: Check domain availability, compare prices across 19 registrars, generate pronounceable names. Use for finding available domains, brand naming, startup naming sessions.
---

# Domain Hunter

Bulk domain availability checker with price comparison.

## When to use
- User asks to find available domain names
- User wants to check if a domain is taken
- User needs domain name suggestions
- User wants to compare domain prices
- Brand naming sessions
- Startup naming

## Setup

Install the MCP server:
```bash
go install github.com/codeatlasdev/domain-hunter/cmd/mcp@latest
```

Add to your MCP config:
```json
{
  "mcpServers": {
    "domh": {
      "command": "domh-mcp"
    }
  }
}
```

## Tools

### check_domain
Check if a single domain is available.
- Input: `domain` (string) - e.g. "coolname.com"
- Returns: availability status, pricing info, buy links

### check_domains
Check multiple domains at once.
- Input: `domains` (string) - comma-separated, e.g. "cool.com,cool.dev,cool.io"
- Returns: array of results with availability and pricing

### check_with_preset
Check a name across a curated TLD set.
- Input: `name` (string), `preset` (string: startup/tech/creative/ecommerce/finance/popular/classic/enterprise/web/trendy/country/brazil)
- Returns: availability across all preset TLDs

### generate_names
Generate pronounceable domain names.
- Input: `length` (number: 3-5), `pattern` (string: CVC/VCV/CVCV/CVCVC/ALL), `tld` (string)
- Returns: list of generated domain names (first 100)

### get_prices
Get registrar prices for a domain.
- Input: `domain` (string)
- Returns: prices from 19 registrars sorted cheapest first, with buy URLs

### list_presets
List all available TLD presets.
- Returns: all 12 presets with their TLD lists
