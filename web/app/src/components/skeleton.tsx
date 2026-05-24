export function SkeletonCard() {
  return (
    <div className="border-[2.5px] border-border/20 bg-white p-4 animate-pulse">
      <div className="flex items-center gap-3">
        <div className="h-5 w-5 bg-border/10" />
        <div className="h-5 w-48 bg-border/10" />
        <div className="h-4 w-16 bg-border/10" />
      </div>
    </div>
  )
}

export function SkeletonList({ count }: { count: number }) {
  return (
    <div className="space-y-2">
      {Array.from({ length: count }, (_, i) => (
        <SkeletonCard key={i} />
      ))}
    </div>
  )
}
