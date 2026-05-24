export interface DomainResult {
  domain: string
  available: boolean
  method: string
  pricing?: {
    domain: string
    tld: string
    cheapest?: { registrar: string; register_price: number; buy_url: string }
    prices: { registrar: string; register_price: number; buy_url: string }[]
  }
}

export interface CheckStats {
  total: number
  available: number
  taken: number
  elapsed_ms: number
}

export function checkStream(
  keyword: string,
  tlds: string[],
  onResult: (r: DomainResult) => void,
  onDone: (s: CheckStats) => void,
  onError: (e: string) => void,
): () => void {
  const params = new URLSearchParams({ keyword, tlds: tlds.join(',') })
  const es = new EventSource(`/api/check/stream?${params}`)

  es.addEventListener('result', (e) => {
    onResult(JSON.parse(e.data))
  })

  es.addEventListener('done', (e) => {
    onDone(JSON.parse(e.data))
    es.close()
  })

  es.onerror = () => {
    onError('Connection lost')
    es.close()
  }

  return () => es.close()
}
