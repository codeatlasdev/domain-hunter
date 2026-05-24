import { useState } from 'react'

const FAQS = [
  {
    q: 'Is domh free?',
    a: 'Yes. CLI is MIT-licensed. Web checker is free, no signup.',
  },
  {
    q: 'How accurate is the data?',
    a: '4-layer verification: DNS (NS + A + MX), RDAP, WHOIS, SSL. False positives are near zero.',
  },
  {
    q: 'What is RDAP?',
    a: 'The modern replacement for WHOIS. Structured, machine-readable domain registration data.',
  },
  {
    q: 'Which registrars?',
    a: 'Namecheap, Porkbun, Cloudflare, GoDaddy, Google, Dynadot, NameSilo, Spaceship, Hostinger, and 10 more.',
  },
  {
    q: 'Can I save my TLD selection?',
    a: 'Yes. Your selection is encoded in the URL. Share the link to share the query.',
  },
]

export function FAQ() {
  const [open, setOpen] = useState<number | null>(null)

  return (
    <div className="space-y-4">
      <h2 className="text-2xl font-black">[+] frequently asked</h2>
      <div className="space-y-2">
        {FAQS.map((faq, i) => (
          <div key={i} className="border-[2.5px] border-border bg-white shadow-brutal-sm">
            <button
              onClick={() => setOpen(open === i ? null : i)}
              className="w-full text-left px-4 py-3 font-bold flex items-center justify-between"
            >
              {faq.q}
              <span className="text-primary font-mono">{open === i ? '[-]' : '[+]'}</span>
            </button>
            {open === i && (
              <div className="px-4 pb-3 text-sm text-slate-600 border-t-[2px] border-border/20 pt-3">
                {faq.a}
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}
