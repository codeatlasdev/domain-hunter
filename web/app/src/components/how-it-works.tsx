export function HowItWorks() {
  const steps = [
    { num: '01', title: 'type a keyword', desc: 'Any word, brand, or acronym. All TLDs get queried in parallel.' },
    { num: '02', title: 'pick your TLDs', desc: 'Use a preset or build a custom set. Saved in the URL.' },
    { num: '03', title: 'get results', desc: 'Results land in real time. Click any available domain to register.' },
  ]

  return (
    <div className="space-y-4">
      <h2 className="text-2xl font-black">[+] how it works</h2>
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
        {steps.map((s) => (
          <div key={s.num} className="border-[2.5px] border-border bg-white p-4 shadow-brutal-sm">
            <div className="text-3xl font-black text-primary/30 mb-2">{s.num}</div>
            <div className="font-bold text-lg mb-1">{s.title}</div>
            <p className="text-sm text-slate-500">{s.desc}</p>
          </div>
        ))}
      </div>
    </div>
  )
}
