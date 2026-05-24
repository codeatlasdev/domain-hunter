import { Globe, Zap, DollarSign, Shield } from 'lucide-react'

interface Props {
  onExample: (keyword: string) => void
}

export function EmptyState({ onExample }: Props) {
  return (
    <div className="space-y-8">
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
        {[
          { icon: Zap, label: '~500/s', desc: 'domains checked' },
          { icon: Globe, label: '1,437', desc: 'TLDs indexed' },
          { icon: DollarSign, label: '19', desc: 'registrars' },
          { icon: Shield, label: '4-layer', desc: 'verification' },
        ].map(({ icon: Icon, label, desc }) => (
          <div key={label} className="border-[2.5px] border-border bg-bg-card p-3 shadow-brutal-sm text-center">
            <Icon size={18} className="mx-auto mb-1 text-primary" />
            <div className="font-black text-lg">{label}</div>
            <div className="text-xs text-slate-500 font-medium">{desc}</div>
          </div>
        ))}
      </div>

      <div className="text-center space-y-3">
        <p className="text-sm text-slate-500 font-medium">try an example:</p>
        <div className="flex flex-wrap justify-center gap-2">
          {['nexo', 'hivo', 'coolname', 'myapp', 'zuno'].map((ex) => (
            <button
              key={ex}
              onClick={() => onExample(ex)}
              className="border-[2.5px] border-border bg-bg-lavender px-3 py-1 text-sm font-bold shadow-brutal-sm hover:translate-x-[-2px] hover:translate-y-[-2px] hover:shadow-brutal-md active:translate-x-[2px] active:translate-y-[2px] active:shadow-none transition-all"
            >
              {ex}
            </button>
          ))}
        </div>
      </div>
    </div>
  )
}
