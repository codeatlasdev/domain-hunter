import { useState } from 'react'
import { Terminal, Globe, DollarSign, Layers, Cpu, Keyboard, Bot, Copy, Check } from 'lucide-react'

function CopyBlock({ code, label }: { code: string; label?: string }) {
  const [copied, setCopied] = useState(false)
  const handleCopy = () => {
    navigator.clipboard.writeText(code)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }
  return (
    <div className="relative group">
      {label && <div className="text-[10px] font-mono uppercase tracking-wider text-slate-500 mb-1">{label}</div>}
      <div className="border-[2.5px] border-border bg-slate-950 text-slate-200 px-4 py-3 pr-12 font-mono text-sm shadow-brutal-sm overflow-x-auto">
        <pre className="whitespace-pre-wrap">{code}</pre>
      </div>
      <button
        onClick={handleCopy}
        className="absolute top-2 right-2 p-1.5 text-slate-500 hover:text-white hover:bg-slate-700 transition-colors rounded"
        title="Copy"
      >
        {copied ? <Check size={14} className="text-available" /> : <Copy size={14} />}
      </button>
    </div>
  )
}

const SECTIONS = [
  {
    id: 'scan',
    icon: Terminal,
    title: 'Scan Domains',
    desc: 'Generate pronounceable names and check availability across multiple TLDs simultaneously.',
    gif: '/docs/scan.gif',
    commands: [
      { label: 'Find 4-letter .com domains', code: 'domh scan --tld com --length 4' },
      { label: 'Check a brand across startup TLDs', code: 'domh scan coolname --preset startup' },
      { label: 'Prefix/suffix combos', code: 'domh scan myapp --prefix get,try --suffix hub,ly --tld com,dev' },
    ],
  },
  {
    id: 'check',
    icon: Globe,
    title: 'Dictionary Mode',
    desc: 'Check a list of your own name ideas against any TLDs.',
    gif: '/docs/check.gif',
    commands: [
      { label: 'Create a names file', code: 'echo -e "nexo\\nhivo\\nzuno\\ntevo" > names.txt' },
      { label: 'Check against multiple TLDs', code: 'domh check names.txt --tld com,dev,io' },
    ],
  },
  {
    id: 'pricing',
    icon: DollarSign,
    title: 'Price Comparison',
    desc: '19 registrars compared instantly. Direct buy links included.',
    gif: '/docs/pricing.gif',
    commands: [
      { label: 'Prices show automatically for available domains', code: 'domh scan zuvotrix --preset startup' },
    ],
  },
  {
    id: 'verification',
    icon: Layers,
    title: '4-Layer Verification',
    desc: 'DNS → RDAP → WHOIS → SSL. No false positives.',
    gif: '/docs/verification.gif',
    commands: [
      { label: 'Show registrar info for taken domains', code: 'domh scan nexo --preset startup --info' },
    ],
  },
  {
    id: 'tui',
    icon: Cpu,
    title: 'Beautiful TUI',
    desc: 'Split-panel dashboard. Progress, stats, available domains, activity log — all at once.',
    gif: '/docs/tui.gif',
    commands: [
      { label: 'Run a scan (TUI shows automatically)', code: 'domh scan --tld com --length 3 --workers 50' },
    ],
  },
  {
    id: 'presets',
    icon: Keyboard,
    title: 'TLD Presets',
    desc: '12 curated presets for different use cases.',
    gif: '/docs/presets.gif',
    commands: [
      { label: 'List all presets', code: 'domh presets' },
      { label: 'Use a preset', code: 'domh scan mybrand --preset tech' },
    ],
  },
  {
    id: 'mcp',
    icon: Bot,
    title: 'AI Agent Integration',
    desc: 'MCP server for Claude, Cursor, VS Code Copilot, and any AI agent.',
    gif: '/docs/mcp.gif',
    commands: [
      { label: 'Install MCP server', code: 'go install github.com/codeatlasdev/domain-hunter/cmd/mcp@latest' },
      { label: 'Or use built-in subcommand', code: 'domh mcp' },
    ],
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

      {/* Install */}
      <section className="space-y-4">
        <h2 className="text-2xl font-black">Install</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="space-y-2">
            <div className="text-sm font-bold text-slate-600">Go (recommended)</div>
            <CopyBlock code="go install github.com/codeatlasdev/domain-hunter@latest" />
          </div>
          <div className="space-y-2">
            <div className="text-sm font-bold text-slate-600">macOS / Linux</div>
            <CopyBlock code="curl -fsSL https://raw.githubusercontent.com/codeatlasdev/domain-hunter/main/install.sh | sh" />
          </div>
        </div>
        <div className="space-y-2">
          <div className="text-sm font-bold text-slate-600">Quick start</div>
          <CopyBlock code="domh" label="Run the interactive wizard" />
        </div>
      </section>

      {/* TOC */}
      <nav className="border-[2.5px] border-border bg-white p-4 shadow-brutal-sm">
        <div className="text-xs font-bold uppercase tracking-wider text-slate-400 mb-3">Sections</div>
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

      {/* Feature sections */}
      {SECTIONS.map((section) => {
        const Icon = section.icon
        return (
          <section key={section.id} id={section.id} className="scroll-mt-24 space-y-5">
            <div className="flex items-center gap-3">
              <div className="border-[2.5px] border-border bg-bg-card p-2 shadow-brutal-sm">
                <Icon size={20} className="text-primary" />
              </div>
              <div>
                <h2 className="text-2xl font-black">{section.title}</h2>
                <p className="text-sm text-slate-500">{section.desc}</p>
              </div>
            </div>

            {/* GIF */}
            <div className="border-[2.5px] border-border shadow-brutal-sm overflow-hidden bg-slate-900">
              <img
                src={`${section.gif}?v=3`}
                alt={`${section.title} demo`}
                className="w-full h-auto block max-h-[300px] object-cover object-top"
              />
            </div>

            {/* Commands */}
            <div className="space-y-3">
              {section.commands.map((cmd, i) => (
                <CopyBlock key={i} code={cmd.code} label={cmd.label} />
              ))}
            </div>
          </section>
        )
      })}

      {/* CLI Reference */}
      <section className="space-y-4">
        <h2 className="text-2xl font-black">CLI Reference</h2>
        <CopyBlock code={`domh                       # Interactive wizard
domh scan [name] [flags]   # Scan domains
domh check <file> [flags]  # Dictionary mode
domh tlds [flags]          # List TLDs
domh presets               # List presets
domh mcp                   # Start MCP server
domh update                # Self-update
domh version               # Version info`} label="Commands" />
        <CopyBlock code={`--tld          TLDs (comma-separated)       [default: com]
--preset       Use preset TLD set
--all          Check ALL 1,437 TLDs
--length       Domain length (3-5)          [default: 3]
--pattern      CVC, VCV, CVCV, ALL          [default: ALL]
--prefix       Prefixes (comma-separated)
--suffix       Suffixes (comma-separated)
--workers      Concurrent workers           [default: 50]
--format       Export: txt,json,csv         [default: txt]
--regex, -r    Regex filter
--batch        Plain output (no TUI)
--yes, -y      Skip confirmations`} label="Scan flags" />
      </section>

      {/* MCP config */}
      <section className="space-y-4">
        <h2 className="text-2xl font-black">MCP Configuration</h2>
        <p className="text-sm text-slate-500">Add to your <code className="bg-slate-100 px-1.5 py-0.5 border border-slate-200 font-mono text-xs">.mcp.json</code>:</p>
        <CopyBlock code={`{
  "mcpServers": {
    "domh": {
      "command": "domh-mcp"
    }
  }
}`} />
        <div className="border-[2.5px] border-border bg-white p-4 shadow-brutal-sm">
          <div className="text-xs font-bold uppercase tracking-wider text-slate-400 mb-3">Available MCP Tools</div>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 text-sm">
            {[
              ['check_domain', 'Check single domain availability'],
              ['check_domains', 'Bulk check (comma-separated)'],
              ['check_with_preset', 'Check name across TLD preset'],
              ['generate_names', 'Generate pronounceable names'],
              ['get_prices', 'Price comparison across 19 registrars'],
              ['list_presets', 'List all TLD presets'],
            ].map(([tool, desc]) => (
              <div key={tool} className="flex gap-2">
                <code className="font-mono font-bold text-primary text-xs">{tool}</code>
                <span className="text-slate-500 text-xs">{desc}</span>
              </div>
            ))}
          </div>
        </div>
      </section>
    </div>
  )
}
