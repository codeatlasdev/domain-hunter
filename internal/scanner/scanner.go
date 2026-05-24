package scanner

import (
	"context"
	"crypto/tls"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Result struct {
	Domain    string
	TLD       string
	Available bool
	Error     bool
	Method    string // "rdap" or "dns"
	Timestamp time.Time
}

type Stats struct {
	Checked   int64
	Available int64
	Errors    int64
	Total     int
	StartTime time.Time
}

type Scanner struct {
	clients  map[string]*http.Client
	workers  int
	stats    Stats
	mu       sync.Mutex
	results  []Result
	resolver *net.Resolver

	OnResult func(Result)
	Done     chan struct{}
}

func New(workers int) *Scanner {
	clients := make(map[string]*http.Client)

	// Separate transport per TLD group to avoid connection pool contention
	for tld := range Providers {
		clients[tld] = &http.Client{
			Timeout: 12 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        200,
				MaxIdleConnsPerHost: 200,
				MaxConnsPerHost:     200,
				IdleConnTimeout:     90 * time.Second,
				TLSHandshakeTimeout: 5 * time.Second,
				TLSClientConfig:     &tls.Config{InsecureSkipVerify: false},
				DisableKeepAlives:   false,
				ForceAttemptHTTP2:   true,
			},
			// Don't follow redirects
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
	}

	return &Scanner{
		clients: clients,
		workers: workers,
		Done:    make(chan struct{}),
		resolver: &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{Timeout: 3 * time.Second}
				return d.DialContext(ctx, network, "1.1.1.1:53")
			},
		},
	}
}

func (s *Scanner) Stats() Stats {
	return Stats{
		Checked:   atomic.LoadInt64(&s.stats.Checked),
		Available: atomic.LoadInt64(&s.stats.Available),
		Errors:    atomic.LoadInt64(&s.stats.Errors),
		Total:     s.stats.Total,
		StartTime: s.stats.StartTime,
	}
}

func (s *Scanner) Results() []Result {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Result, len(s.results))
	copy(out, s.results)
	return out
}

func (s *Scanner) Run(domains []string) {
	s.stats.Total = len(domains)
	s.stats.StartTime = time.Now()

	// Shuffle to distribute load across TLDs evenly
	shuffled := make([]string, len(domains))
	copy(shuffled, domains)
	rand.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })

	work := make(chan string, s.workers*4)
	var wg sync.WaitGroup

	for i := 0; i < s.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for domain := range work {
				r := s.check(domain)
				atomic.AddInt64(&s.stats.Checked, 1)
				if r.Available {
					atomic.AddInt64(&s.stats.Available, 1)
				}
				if r.Error {
					atomic.AddInt64(&s.stats.Errors, 1)
				}
				s.mu.Lock()
				s.results = append(s.results, r)
				s.mu.Unlock()
				if s.OnResult != nil {
					s.OnResult(r)
				}
			}
		}()
	}

	go func() {
		for _, d := range shuffled {
			work <- d
		}
		close(work)
		wg.Wait()
		close(s.Done)
	}()
}

func (s *Scanner) check(domain string) Result {
	parts := strings.SplitN(domain, ".", 2)
	if len(parts) != 2 {
		return Result{Domain: domain, Error: true, Timestamp: time.Now()}
	}
	tld := parts[1]

	baseURL, hasRDAP := Providers[tld]

	// Strategy 1: RDAP (preferred — authoritative)
	if hasRDAP {
		r := s.checkRDAP(domain, tld, baseURL)
		if !r.Error {
			return r
		}
		// RDAP failed — fallback to DNS for unreliable TLDs
	}

	// Strategy 2: DNS NXDOMAIN check (fallback)
	return s.checkDNS(domain, tld)
}

func (s *Scanner) checkRDAP(domain, tld, baseURL string) Result {
	client := s.clients[tld]
	url := baseURL + domain

	var resp *http.Response
	var err error

	for attempt := 0; attempt < 3; attempt++ {
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Accept", "application/rdap+json")

		resp, err = client.Do(req)
		if err == nil {
			break
		}

		// Exponential backoff: 200ms, 500ms, 1s
		backoff := time.Duration(200*(1<<attempt)) * time.Millisecond
		time.Sleep(backoff)
	}

	if err != nil {
		return Result{Domain: domain, TLD: tld, Error: true, Method: "rdap", Timestamp: time.Now()}
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 404, 400:
		// Not found = available
		return Result{Domain: domain, TLD: tld, Available: true, Method: "rdap", Timestamp: time.Now()}
	case 200:
		// Found = taken
		return Result{Domain: domain, TLD: tld, Available: false, Method: "rdap", Timestamp: time.Now()}
	case 429:
		// Rate limited — sleep and mark as error for retry
		time.Sleep(2 * time.Second)
		return Result{Domain: domain, TLD: tld, Error: true, Method: "rdap", Timestamp: time.Now()}
	default:
		// Unexpected status — treat as error
		return Result{Domain: domain, TLD: tld, Error: true, Method: "rdap", Timestamp: time.Now()}
	}
}

func (s *Scanner) checkDNS(domain, tld string) Result {
	// Check if domain has any DNS records (NS, A, AAAA)
	// If NXDOMAIN → likely available (not 100% authoritative but good signal)
	_, err := net.LookupNS(domain)
	if err != nil {
		// Check if it's NXDOMAIN (not found) vs network error
		if dnsErr, ok := err.(*net.DNSError); ok && dnsErr.IsNotFound {
			// Double-check with A record
			_, err2 := net.LookupHost(domain)
			if err2 != nil {
				if dnsErr2, ok2 := err2.(*net.DNSError); ok2 && dnsErr2.IsNotFound {
					return Result{Domain: domain, TLD: tld, Available: true, Method: "dns", Timestamp: time.Now()}
				}
			}
			return Result{Domain: domain, TLD: tld, Available: true, Method: "dns", Timestamp: time.Now()}
		}
		// Network error
		return Result{Domain: domain, TLD: tld, Error: true, Method: "dns", Timestamp: time.Now()}
	}

	// Has NS records = registered
	return Result{Domain: domain, TLD: tld, Available: false, Method: "dns", Timestamp: time.Now()}
}

// GenerateAll creates domains for given TLDs, length, and pattern
func GenerateAll(tlds []string, length int, pattern string) []string {
	var all []string
	names := Generate(length, pattern)
	for _, tld := range tlds {
		for _, name := range names {
			all = append(all, fmt.Sprintf("%s.%s", name, tld))
		}
	}
	return all
}
