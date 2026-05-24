import { Outlet, Link } from '@tanstack/react-router'

export function Layout() {
  return (
    <div className="min-h-screen flex flex-col">
      <header className="border-b-[2.5px] border-border bg-white px-6 py-4 flex items-center justify-between">
        <div className="flex items-center gap-6">
          <Link to="/" className="text-xl font-black text-primary tracking-tight">
            ◆ domh
          </Link>
          <nav className="hidden sm:flex items-center gap-1 text-sm font-bold">
            <Link
              to="/docs"
              className="px-2 py-1 hover:bg-bg-card transition-colors"
              activeProps={{ className: 'bg-bg-card' }}
            >
              Docs
            </Link>
            <Link
              to="/tlds"
              className="px-2 py-1 hover:bg-bg-card transition-colors"
              activeProps={{ className: 'bg-bg-card' }}
            >
              TLDs
            </Link>
          </nav>
        </div>
        <a
          href="https://github.com/codeatlasdev/domain-hunter"
          target="_blank"
          rel="noopener noreferrer"
          className="flex items-center gap-2 border-[2.5px] border-border px-3 py-1.5 font-bold text-sm shadow-brutal-sm hover:translate-x-[-2px] hover:translate-y-[-2px] hover:shadow-brutal-md transition-all"
        >
          [★] GitHub
        </a>
      </header>

      <main className="flex-1 max-w-4xl mx-auto w-full px-6 py-10">
        <Outlet />
      </main>

      <footer className="border-t-[2.5px] border-border bg-white px-6 py-8">
        <div className="max-w-4xl mx-auto flex flex-wrap gap-8 justify-between text-sm">
          <div>
            <div className="font-black text-primary mb-2">◆ domh</div>
            <p className="text-slate-500 text-xs max-w-[200px]">
              Bulk domain availability checker across 1,437 TLDs. Open source.
            </p>
          </div>
          <div>
            <div className="font-bold mb-2">Product</div>
            <div className="space-y-1 text-slate-500">
              <Link to="/" className="block hover:text-slate-800">Checker</Link>
              <Link to="/docs" className="block hover:text-slate-800">Docs</Link>
              <Link to="/tlds" className="block hover:text-slate-800">TLDs</Link>
            </div>
          </div>
          <div>
            <div className="font-bold mb-2">Resources</div>
            <div className="space-y-1 text-slate-500">
              <a href="https://github.com/codeatlasdev/domain-hunter" className="block hover:text-slate-800">GitHub</a>
              <a href="https://github.com/codeatlasdev/domain-hunter/releases" className="block hover:text-slate-800">Releases</a>
            </div>
          </div>
          <div>
            <div className="font-bold mb-2">Company</div>
            <div className="space-y-1 text-slate-500">
              <a href="https://codeatlas.com.br" className="block hover:text-slate-800">CodeAtlas</a>
            </div>
          </div>
        </div>
        <div className="max-w-4xl mx-auto mt-6 pt-4 border-t border-slate-200 text-xs text-slate-400 text-center">
          © 2026 domh · <a href="https://codeatlas.com.br" className="hover:underline">CodeAtlas</a>
        </div>
      </footer>
    </div>
  )
}
