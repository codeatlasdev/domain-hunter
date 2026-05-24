import { useState, useRef, useCallback, useEffect } from 'react'
import { toast } from 'sonner'
import { SearchInput } from '../components/search-input'
import { PresetPills } from '../components/preset-pills'
import { ResultList } from '../components/result-list'
import { StatsBar } from '../components/stats-bar'
import { EmptyState } from '../components/empty-state'
import { SkeletonList } from '../components/skeleton'
import { HowItWorks } from '../components/how-it-works'
import { FAQ } from '../components/faq'
import { checkStream, type DomainResult, type CheckStats } from '../lib/api'

export function Home() {
  const [keyword, setKeyword] = useState('')
  const [selectedPreset, setSelectedPreset] = useState<string | null>('Startup Picks')
  const [tlds, setTlds] = useState<string[]>(['com', 'org', 'io', 'ai', 'tech', 'app', 'dev', 'xyz'])
  const [results, setResults] = useState<DomainResult[]>([])
  const [stats, setStats] = useState<CheckStats | null>(null)
  const [checking, setChecking] = useState(false)
  const [checked, setChecked] = useState(0)
  const [total, setTotal] = useState(0)
  const [hasSearched, setHasSearched] = useState(false)
  const closeRef = useRef<(() => void) | null>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    return () => { closeRef.current?.() }
  }, [])

  // Keyboard shortcut: "/" to focus search
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (e.key === '/' && document.activeElement?.tagName !== 'INPUT') {
        e.preventDefault()
        inputRef.current?.focus()
      }
    }
    document.addEventListener('keydown', handler)
    return () => document.removeEventListener('keydown', handler)
  }, [])

  const handlePreset = useCallback((name: string, presetTlds: string[]) => {
    setSelectedPreset(name)
    setTlds(presetTlds)
  }, [])

  const handleToggleTld = useCallback((tld: string) => {
    setTlds((prev) => prev.filter((t) => t !== tld))
    setSelectedPreset(null)
  }, [])

  const handleSubmit = useCallback(() => {
    if (!keyword.trim() || tlds.length === 0) return

    setResults([])
    setStats(null)
    setChecking(true)
    setChecked(0)
    setTotal(tlds.length)
    setHasSearched(true)

    closeRef.current?.()
    closeRef.current = checkStream(
      keyword.trim(),
      tlds,
      (r) => {
        setResults((prev) => [...prev, r])
        setChecked((c) => c + 1)
      },
      (s) => {
        setStats(s)
        setChecking(false)
      },
      (err) => {
        toast.error(err)
        setChecking(false)
      },
    )
  }, [keyword, tlds])

  const handleExample = useCallback((ex: string) => {
    setKeyword(ex)
    inputRef.current?.focus()
  }, [])

  const remainingSkeletons = checking ? Math.max(0, total - checked) : 0

  return (
    <div className="space-y-8">
      <div className="text-center space-y-2">
        <p className="text-xs font-mono uppercase tracking-widest text-slate-400">domain availability checker</p>
        <h1 className="text-5xl font-black tracking-tight">domh</h1>
        <p className="text-slate-500 font-medium">check any keyword across every TLD.</p>
      </div>

      <SearchInput
        ref={inputRef}
        value={keyword}
        onChange={setKeyword}
        onSubmit={handleSubmit}
        disabled={checking}
      />

      <PresetPills
        selected={selectedPreset}
        tlds={tlds}
        onSelect={handlePreset}
        onToggleTld={handleToggleTld}
      />

      {hasSearched && (
        <>
          <StatsBar stats={stats} checking={checking} checked={checked} total={total} />
          <ResultList results={results} total={total} checked={checked} checking={checking} />
          {remainingSkeletons > 0 && <SkeletonList count={remainingSkeletons} />}
        </>
      )}

      {!hasSearched && (
        <>
          <EmptyState onExample={handleExample} />
          <HowItWorks />
          <FAQ />
        </>
      )}
    </div>
  )
}
