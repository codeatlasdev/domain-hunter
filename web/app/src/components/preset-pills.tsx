const PRESETS: Record<string, string[]> = {
  'Startup Picks': ['com', 'org', 'io', 'ai', 'tech', 'app', 'dev', 'xyz'],
  'Create': ['design', 'art', 'studio', 'media', 'gallery', 'ink'],
  'Publish': ['blog', 'news', 'press', 'media', 'page', 'site'],
  'Sell': ['shop', 'store', 'market', 'sale', 'deals', 'buy'],
  'Regional': ['us', 'uk', 'de', 'fr', 'ca', 'au', 'br', 'in', 'nl', 'jp'],
  'Brazil': ['com.br', 'net.br', 'org.br', 'app.br', 'dev.br'],
  'All gTLDs': ['com', 'net', 'org', 'io', 'ai', 'app', 'dev', 'co', 'xyz', 'me', 'tech', 'cloud', 'page', 'tools', 'run'],
}

interface Props {
  selected: string | null
  tlds: string[]
  onSelect: (name: string, tlds: string[]) => void
  onToggleTld: (tld: string) => void
}

export function PresetPills({ selected, tlds, onSelect, onToggleTld }: Props) {
  return (
    <div className="space-y-3">
      {/* Preset buttons */}
      <div className="flex flex-wrap gap-2">
        {Object.entries(PRESETS).map(([name, presetTlds]) => (
          <button
            key={name}
            onClick={() => onSelect(name, presetTlds)}
            className={`border-[2.5px] border-border px-3 py-1.5 text-sm font-bold transition-all
              ${selected === name
                ? 'bg-slate-900 text-white shadow-brutal-sm'
                : 'bg-white shadow-brutal-sm hover:translate-x-[-2px] hover:translate-y-[-2px] hover:shadow-brutal-md'
              }`}
          >
            {name}
          </button>
        ))}
      </div>

      {/* Selected TLD pills */}
      {tlds.length > 0 && (
        <div className="flex flex-wrap gap-1.5">
          {tlds.map((tld) => (
            <button
              key={tld}
              onClick={() => onToggleTld(tld)}
              className="border-[2px] border-border/60 bg-white px-2 py-0.5 text-xs font-mono font-bold hover:bg-red-50 hover:border-red-300 hover:line-through transition-all"
            >
              .{tld}
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
