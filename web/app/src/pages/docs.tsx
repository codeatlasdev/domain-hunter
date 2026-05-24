import { Terminal, Globe, Zap, DollarSign, Layers, Cpu, Keyboard, Bot } from 'lucide-react'

const SECTIONS = [
  {
    id: 'quick-start',
    icon: Zap,
    title: 'Quick Start',
    desc: 'Install and run your first scan in under 10 seconds.',
    gif: '/docs/quick-start.gif',
    code: `# Install
go install github.com/codeatlasdev/domain-hunter@latest

# Run the interactive wizard
domh`,
  },
  {
    id: 'scan',
    icon: Terminal,
    title: 'Scan Domains',
    desc: 'Generate pronounceable names and check availability across multiple TLDs simultaneously.',
    gif: '/docs/scan.gif',
    code: `# Find 4-letter .com domains
domh scan --tld com --length 4

# Check a brand across startup TLDs
domh scan coolname --preset startup

# Prefix/suffix combos
domh scan myapp --prefix get,try --suffix hub,ly --tld com,dev`,
  },
  {
    id: 'check',
    icon: Globe,
    title: 'Dictionary Mode',
    desc: 'Check a list of your own name ideas against any TLDs.',
    gif: '/docs/check.gif',
    code: `# Create a names file
echo -e "nexo\\nhivo\\nzuno\\ntevo" > names.txt

# Check against multiple TLDs
domh check names.txt --tld com,dev,io`,
  },
  {
    id: 'pricing',
    icon: DollarSign,
    title: 'Price Comparison',
    desc: '19 registrars compared instantly. Direct buy links included.',
    gif: '/docs/pricing.gif',
    code: `# Prices are shown automatically for available domains
# Includes: Namecheap, Porkbun, Cloudflare, GoDaddy,
# Google, Dynadot, NameSilo, Spaceship, Hostinger,
# IONOS, OVH, Gandi, Hover, Epik, Name.com,
# Registro.br, 101domain, INWX, Dreamhost`,
  },
  {
    id: 'verification',
    icon: Layers,
    title: '4-Layer Verification',
    desc: 'DNS → RDAP → WHOIS → SSL. No false positives.',
    gif: '/docs/verification.gif',
    code: `# Layer 1: DNS (NS → A → MX) — fast, no rate limit
# Layer 2: RDAP — authoritative confirmation
# Layer 3: WHOIS — multi-server fallback
# Layer 4: SSL — last resort for edge cases

# ~500 domains/second with 50 workers
domh scan --workers 50 --tld com --length 3`,
  },
  {
    id: 'tui',
    icon: Cpu,
    title: 'Beautiful TUI',
    desc: 'Split-panel dashboard that uses 100% of your terminal. Responsive.',
    gif: '/docs/tui.gif',
    code: `# The TUI shows automatically during scans
# Split panels: available domains + activity log
# Live stats: progress, rate, ETA
# Responsive to terminal size`,
  },
  {
    id: 'presets',
    icon: Keyboard,
    title: 'TLD Presets',
    desc: '12 curated presets for different use cases.',
    gif: '/docs/presets.gif',
    code: `# List all presets
domh presets

# Built-in presets:
# startup  → com, org, io, ai, tech, app, dev, xyz
# tech     → io, ai, app, dev, tech, cloud, software
# creative → design, art, studio, media, photography
# ecommerce → shop, store, market, sale, deals
# finance  → finance, capital, fund, money, bank
# brazil   → com.br, net.br, org.br, app.br, dev.br`,
  },
  {
    id: 'mcp',
    icon: Bot,
    title: 'AI Agent Integration',
    desc: 'MCP server for Claude, Cursor, VS Code Copilot, and any AI agent.',
    gif: '/docs/mcp.gif',
    code: `# Install MCP server
go install github.com/codeatlasdev/domain-hunter/cmd/mcp@latest

# Add to .mcp.json
{
  "mcpServers": {
    "domh": { "command": "domh-mcp" }
  }
}

# Available tools: check_domain, check_domains,
# check_with_preset, generate_names, get_prices, list_presets`,
  },
]

export function Docs() {
  return (
    <div className="space-y-12">
      {/* Hero */}
      <div className="space-y-4">
        <h1 className="text-4xl font-black tracking-tight">Documentation</h1>
        <p className="text-slate-500 font-medium max-w-xl">
          Everything you need to find your next domain. CLI, web checker, and AI integration.
        </p>
      </div>

      {/* TOC */}
      <nav className="border-[2.5px] border-border bg-white p-4 shadow-brutal-sm">
        <div className="text-xs font-bold uppercase tracking-wider text-slate-400 mb-3">On this page</div>
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-2">
          {SECTIONS.map((s) => (
            <a
              key={s.id}
              href={`#${s.id}`}
              className="text-sm font-bold text-slate-600 hover:text-primary transition-colors"
            >
              {s.title}
            </a>
          ))}
        </div>
      </nav>

      {/* Install banner */}
      <div className="border-[2.5px] border-border bg-slate-900 text-white p-6 shadow-brutal-md">
        <div className="text-xs font-mono uppercase tracking-wider text-slate-400 mb-2">Install</div>
        <code className="text-lg font-bold">go install github.com/codeatlasdev/domain-hunter@latest</code>
        <div className="mt-3 text-sm text-slate-400">
          Or: <code className="text-slate-300">curl -fsSL https://raw.githubusercontent.com/codeatlasdev/domain-hunter/main/install.sh | sh</code>
        </div>
      </div>

      {/* Feature sections */}
      {SECTIONS.map((section, i) => {
        const Icon = section.icon
        const isEven = i % 2 === 0
        return (
          <section key={section.id} id={section.id} className="scroll-mt-24 space-y-4">
            {/* Section header */}
            <div className="flex items-center gap-3">
              <div className="border-[2.5px] border-border bg-bg-card p-2 shadow-brutal-sm">
                <Icon size={20} className="text-primary" />
              </div>
              <div>
                <h2 className="text-2xl font-black">{section.title}</h2>
                <p className="text-sm text-slate-500">{section.desc}</p>
              </div>
            </div>

            {/* Content: GIF + Code */}
            <div className={`grid grid-cols-1 lg:grid-cols-2 gap-4 ${isEven ? '' : 'lg:direction-rtl'}`}>
              {/* GIF */}
              <div className="border-[2.5px] border-border bg-slate-900 shadow-brutal-sm overflow-hidden">
                <img
                  src={section.gif}
                  alt={`${section.title} demo`}
                  className="w-full h-auto"
                  loading="lazy"
                />
              </div>

              {/* Code block */}
              <div className="border-[2.5px] border-border bg-slate-950 text-slate-300 p-4 shadow-brutal-sm overflow-x-auto">
                <div className="text-[10px] font-mono uppercase tracking-wider text-slate-500 mb-3">terminal</div>
                <pre className="text-sm font-mono leading-relaxed whitespace-pre-wrap">{section.code}</pre>
              </div>
            </div>
          </section>
        )
      })}

      {/* CLI reference */}
      <section className="space-y-4">
        <h2 className="text-2xl font-black">CLI Reference</h2>
        <div className="border-[2.5px] border-border bg-slate-950 text-slate-300 p-6 shadow-brutal-sm overflow-x-auto">
          <pre className="text-xs font-mono leading-relaxed">{`◆ domh — bulk domain availability checker

Usage:
  domh                       Interactive wizard
  domh scan [name] [flags]   Scan domains
  domh check <file> [flags]  Dictionary mode
  domh tlds [flags]          List TLDs
  domh presets               List presets
  domh mcp                   Start MCP server (stdio)
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
  --yes, -y      Skip confirmations`}</pre>
        </div>
      </section>

      {/* Self-update */}
      <section className="border-[2.5px] border-border bg-bg-card p-6 shadow-brutal-sm">
        <h3 className="font-black text-lg mb-2">Self-Update</h3>
        <code className="font-mono text-sm bg-white border border-border/40 px-2 py-1">domh update</code>
        <p className="text-sm text-slate-500 mt-2">Updates to the latest release automatically. No re-download needed.</p>
      </section>
    </div>
  )
}
