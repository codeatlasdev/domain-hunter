import { useState } from 'react'
import type { DomainResult } from '../lib/api'
import { ResultCard } from './result-card'

interface Props {
  results: DomainResult[]
  total: number
  checked: number
  checking: boolean
}

export function ResultList({ results, total, checked, checking }: Props) {
  const [sortByStatus, setSortByStatus] = useState(false)

  if (results.length === 0 && !checking) return null

  const availableCount = results.filter((r) => r.available).length

  const sorted = sortByStatus
    ? [...results].sort((a, b) => {
        if (a.available && !b.available) return -1
        if (!a.available && b.available) return 1
        return 0
      })
    : results

  return (
    <div className="space-y-3">
      {/* Results header */}
      <div className="flex items-center justify-between flex-wrap gap-2">
        <div className="font-mono text-sm font-medium text-slate-600">
          <span className="font-bold text-slate-800">[+] {total} selected</span>
          {' · '}
          <span>completed {checked}/{total}</span>
          {availableCount > 0 && (
            <>
              {' · '}
              <span className="text-available font-bold">{availableCount} available</span>
            </>
          )}
        </div>
        <button
          onClick={() => setSortByStatus((s) => !s)}
          className="border-[2.5px] border-border px-2.5 py-1 text-xs font-bold shadow-brutal-sm hover:translate-x-[-1px] hover:translate-y-[-1px] hover:shadow-brutal-md active:translate-x-[1px] active:translate-y-[1px] active:shadow-none transition-all"
        >
          [s] sort by status
        </button>
      </div>

      {/* Result cards - list layout */}
      <div className="space-y-2">
        {sorted.map((r, i) => (
          <ResultCard key={r.domain} result={r} index={i} />
        ))}
      </div>
    </div>
  )
}
