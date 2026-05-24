import { useState } from 'react'

const FAQS = [
  {
    q: 'Is domh free?',
    a: 'Yes. The CLI is free and open source (MIT). The web checker is free with no signup required.',
  },
  {
    q: 'How accurate is the availability data?',
    a: 'We use a 4-layer verification: DNS (NS + A + MX), RDAP, WHOIS, and SSL. False positives are extremely rare.',
  },
  {
    q: 'What is RDAP?',
    a: 'RDAP (Registration Data Access Protocol) is the modern replacement for WHOIS. It provides structured, machine-readable domain registration data.',
  },
  {
    q: 'Which registrars do you link to?',
    a: 'We link to 19 registrars including Namecheap, Porkbun, Cloudflare, GoDaddy, Google Domains, Dynadot, NameSilo, Spaceship, and more.',
  },
  {
    q: 'Can I save my TLD selection?',
    a: 'Your TLD selection is saved in the URL. Share the URL to share your exact query and TLD set.',
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
