import { forwardRef } from 'react'
import { Search } from 'lucide-react'

interface Props {
  value: string
  onChange: (v: string) => void
  onSubmit: () => void
  disabled?: boolean
}

export const SearchInput = forwardRef<HTMLInputElement, Props>(
  ({ value, onChange, onSubmit, disabled }, ref) => {
    return (
      <form
        onSubmit={(e) => { e.preventDefault(); onSubmit() }}
        className="flex gap-0"
      >
        <div className="relative flex-1">
          <span className="absolute left-4 top-1/2 -translate-y-1/2 text-slate-400 font-mono text-sm">$</span>
          <input
            ref={ref}
            type="text"
            value={value}
            onChange={(e) => onChange(e.target.value)}
            placeholder="enter a keyword or paste a domain"
            disabled={disabled}
            autoFocus
            className="w-full border-[2.5px] border-border border-r-0 pl-9 pr-4 py-3.5 text-lg font-medium shadow-brutal-sm focus:outline-none focus:shadow-brutal-md transition-shadow disabled:opacity-50"
          />
          <kbd className="absolute right-3 top-1/2 -translate-y-1/2 text-xs text-slate-300 font-mono border border-slate-200 px-1.5 py-0.5 hidden sm:inline">/</kbd>
        </div>
        <button
          type="submit"
          disabled={disabled || !value.trim()}
          className="border-[2.5px] border-border bg-primary text-white px-6 py-3.5 font-bold tracking-wider shadow-brutal-sm hover:translate-x-[-2px] hover:translate-y-[-2px] hover:shadow-brutal-md active:translate-x-[2px] active:translate-y-[2px] active:shadow-none transition-all disabled:opacity-50 disabled:hover:translate-x-0 disabled:hover:translate-y-0 flex items-center gap-2"
        >
          {disabled ? 'checking...' : <><Search size={18} /> <span className="hidden sm:inline">check</span></>}
        </button>
      </form>
    )
  }
)
