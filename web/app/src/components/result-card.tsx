import { ExternalLink } from 'lucide-react'
import type { DomainResult } from '../lib/api'

const REGISTRAR_LINKS: Record<string, (d: string) => string> = {
  namecheap: (d) => `https://www.namecheap.com/domains/registration/results/?domain=${d}`,
  porkbun: (d) => `https://porkbun.com/checkout/search?q=${d}`,
  godaddy: (d) => `https://www.godaddy.com/domainsearch/find?domainToCheck=${d}`,
}

function getTldType(tld: string): string {
  // 2-letter TLDs are ccTLDs, rest are gTLDs (simplified heuristic)
  const base = tld.split('.')[0]
  return base.length === 2 ? 'ccTLD' : 'gTLD'
}

interface Props {
  result: DomainResult
  index: number
}

export function ResultCard({ result, index }: Props) {
  const parts = result.domain.split('.')
  const tld = parts.slice(1).join('.')
  const tldType = getTldType(tld)
  const stagger = `${Math.min(index * 40, 400)}ms`

  if (result.available) {
    return (
      <div
        className="border-[2.5px] border-border bg-bg-mint/50 p-4 shadow-brutal-sm animate-[fadeSlideIn_0.3s_ease-out_both]"
        style={{ animationDelay: stagger }}
      >
        <div className="flex items-center justify-between flex-wrap gap-2">
          <div className="flex items-center gap-2">
            <span className="text-available font-black">[✓]</span>
            <span className="font-bold text-lg">{result.domain}</span>
            <span className="text-available font-bold text-sm">available</span>
          </div>
          <span className="text-xs text-slate-500 font-mono">{tldType}</span>
        </div>
        <div className="mt-3 flex flex-wrap gap-2">
          {Object.entries(REGISTRAR_LINKS).map(([name, urlFn]) => (
            <a
              key={name}
              href={urlFn(result.domain)}
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-1 border-[2.5px] border-border bg-slate-900 text-white text-xs font-bold px-3 py-1.5 shadow-brutal-sm hover:translate-x-[-1px] hover:translate-y-[-1px] hover:shadow-brutal-md active:translate-x-[1px] active:translate-y-[1px] active:shadow-none transition-all"
            >
              {name} <ExternalLink size={10} />
            </a>
          ))}
        </div>
      </div>
    )
  }

  return (
    <div
      className="border-[2.5px] border-border/40 bg-white p-4 animate-[fadeSlideIn_0.3s_ease-out_both]"
      style={{ animationDelay: stagger }}
    >
      <div className="flex items-center justify-between flex-wrap gap-2">
        <div className="flex items-center gap-2">
          <span className="text-slate-400 font-black">[x]</span>
          <a
            href={`https://${result.domain}/`}
            target="_blank"
            rel="noopener noreferrer"
            className="font-bold text-lg underline decoration-slate-300 hover:decoration-slate-500 transition-colors"
          >
            {result.domain}
          </a>
          <span className="text-slate-400 font-bold text-sm">taken</span>
        </div>
        <span className="text-xs text-slate-500 font-mono">{tldType}</span>
      </div>
    </div>
  )
}
