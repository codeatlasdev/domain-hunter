import { useState, useEffect } from 'react'
import { Search } from 'lucide-react'

interface TLD {
  name: string
  type: string
  has_rdap: boolean
}

export function TLDs() {
  const [tlds, setTlds] = useState<TLD[]>([])
  const [search, setSearch] = useState('')
  const [tab, setTab] = useState<'all' | 'gtld' | 'cctld'>('all')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetch('/api/tlds')
      .then((r) => r.json())
      .then((data) => {
        setTlds(data.tlds || [])
        setLoading(false)
      })
      .catch(() => setLoading(false))
  }, [])

  const filtered = tlds.filter((t) => {
    if (tab === 'gtld' && t.type !== 'generic') return false
    if (tab === 'cctld' && t.type !== 'country-code') return false
    if (search) {
      const q = search.toLowerCase().replace(/^\./, '')
      if (!t.name.includes(q)) return false
    }
    return true
  })

  // Group by first letter
  const grouped = filtered.reduce<Record<string, TLD[]>>((acc, t) => {
    const letter = t.name[0].toUpperCase()
    if (!acc[letter]) acc[letter] = []
    acc[letter].push(t)
    return acc
  }, {})

  const tabs = [
    { id: 'all' as const, label: 'All', count: tlds.length },
    { id: 'gtld' as const, label: 'gTLD', count: tlds.filter((t) => t.type === 'generic').length },
    { id: 'cctld' as const, label: 'ccTLD', count: tlds.filter((t) => t.type === 'country-code').length },
  ]

  return (
    <div className="space-y-6">
      <h1 className="text-4xl font-black tracking-tight">All TLDs</h1>

      {/* Search */}
      <div className="relative">
        <Search size={18} className="absolute left-4 top-1/2 -translate-y-1/2 text-slate-400" />
        <input
          type="text"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Search TLD..."
          className="w-full border-[2.5px] border-border pl-11 pr-4 py-3 text-sm font-medium shadow-brutal-sm focus:outline-none focus:shadow-brutal-md transition-shadow"
        />
      </div>

      {/* Tabs */}
      <div className="flex gap-2">
        {tabs.map((t) => (
          <button
            key={t.id}
            onClick={() => setTab(t.id)}
            className={`border-[2.5px] border-border px-3 py-1.5 text-sm font-bold transition-all
              ${tab === t.id
                ? 'bg-slate-900 text-white shadow-brutal-sm'
                : 'bg-white shadow-brutal-sm hover:translate-x-[-1px] hover:translate-y-[-1px] hover:shadow-brutal-md'
              }`}
          >
            {t.label} <span className="text-xs opacity-70">({t.count})</span>
          </button>
        ))}
      </div>

      {/* TLD list */}
      {loading ? (
        <div className="animate-pulse space-y-4">
          {Array.from({ length: 5 }, (_, i) => (
            <div key={i} className="h-6 w-48 bg-border/10" />
          ))}
        </div>
      ) : (
        <div className="space-y-6">
          {Object.entries(grouped).sort(([a], [b]) => a.localeCompare(b)).map(([letter, items]) => (
            <div key={letter}>
              <h2 className="text-lg font-black text-primary mb-2">[{letter}]</h2>
              <div className="flex flex-wrap gap-1.5">
                {items.map((t) => (
                  <span
                    key={t.name}
                    className={`border-[2px] px-2 py-0.5 text-xs font-mono font-bold
                      ${t.type === 'country-code'
                        ? 'border-amber-300 bg-amber-50'
                        : 'border-border/40 bg-white'
                      }
                      ${t.has_rdap ? '' : 'opacity-50'}
                    `}
                  >
                    .{t.name}
                    <span className="text-[10px] ml-1 text-slate-400">
                      {t.type === 'country-code' ? 'cc' : 'g'}
                    </span>
                  </span>
                ))}
              </div>
            </div>
          ))}
        </div>
      )}

      <p className="text-xs text-slate-400 font-mono">
        {filtered.length} TLDs shown · data from IANA
      </p>
    </div>
  )
}
