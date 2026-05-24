interface Props {
  stats: { elapsed_ms: number } | null
  checking: boolean
  checked: number
  total: number
}

export function StatsBar({ stats, checking, checked, total }: Props) {
  if (!checking && !stats) return null

  const pct = total > 0 ? Math.round((checked / total) * 100) : 0

  if (checking) {
    return (
      <div className="border-[2.5px] border-border bg-white shadow-brutal-sm overflow-hidden">
        <div className="h-1 bg-border/10">
          <div
            className="h-full bg-primary transition-all duration-300 ease-out"
            style={{ width: `${pct}%` }}
          />
        </div>
      </div>
    )
  }

  return null
}
