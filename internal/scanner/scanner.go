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
	Domain     string
	TLD        string
	Available  bool
	Error      bool
	Method     string   // primary method that determined result
	Signatures []string // all methods that confirmed: "DNS_NS", "DNS_A", "DNS_MX", "RDAP", "WHOIS", "SSL"
	Timestamp  time.Time
}

type Stats struct {
	Checked   int64
	Available int64
	Errors    int64
	Total     int
	StartTime time.Time
}

// Public DNS resolvers — distributed across providers to avoid rate limiting.
// Each resolver handles a fraction of the load.
var dnsResolvers = []string{
	"1.1.1.1:53",         // Cloudflare
	"1.0.0.1:53",         // Cloudflare secondary
	"8.8.8.8:53",         // Google
	"8.8.4.4:53",         // Google secondary
	"9.9.9.9:53",         // Quad9
	"149.112.112.112:53", // Quad9 secondary
	"208.67.222.222:53",  // OpenDNS
	"208.67.220.220:53",  // OpenDNS secondary
	"76.76.2.0:53",       // ControlD
	"76.76.10.0:53",      // ControlD secondary
	"94.140.14.14:53",    // AdGuard
	"94.140.15.15:53",    // AdGuard secondary
	"185.228.168.9:53",   // CleanBrowsing
	"185.228.169.9:53",   // CleanBrowsing secondary
	"76.223.122.150:53",  // Alternate DNS
	"198.101.242.72:53",  // Alternate DNS secondary
}

type Scanner struct {
	httpClient *http.Client
	resolvers  []*net.Resolver
	workers    int
	delay      time.Duration
	stats      Stats
	mu         sync.Mutex
	results    []Result

	OnResult func(Result)
	Done     chan struct{}
}

func New(workers int) *Scanner {
	return NewWithDelay(workers, 0)
}

func NewWithDelay(workers int, delay time.Duration) *Scanner {
	// Create a resolver pool — each resolver uses a different upstream
	resolvers := make([]*net.Resolver, len(dnsResolvers))
	for i, addr := range dnsResolvers {
		resolverAddr := addr // capture
		resolvers[i] = &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{Timeout: 3 * time.Second}
				return d.DialContext(ctx, "udp", resolverAddr)
			},
		}
	}

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        300,
			MaxIdleConnsPerHost: 100,
			MaxConnsPerHost:     100,
			IdleConnTimeout:     90 * time.Second,
			TLSHandshakeTimeout: 5 * time.Second,
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: false},
			ForceAttemptHTTP2:   true,
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return &Scanner{
		httpClient: httpClient,
		resolvers:  resolvers,
		workers:    workers,
		delay:      delay,
		Done:       make(chan struct{}),
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

	// Shuffle to distribute load
	shuffled := make([]string, len(domains))
	copy(shuffled, domains)
	rand.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })

	work := make(chan string, s.workers*4)
	var wg sync.WaitGroup

	for i := 0; i < s.workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for domain := range work {
				r := s.check(domain, workerID)
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
				if s.delay > 0 {
					time.Sleep(s.delay)
				}
			}
		}(i)
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

// check uses a multi-phase strategy:
// Phase 1: DNS (NS + A + MX) — fast, no rate limit
// Phase 2: RDAP confirmation — for candidates
// Phase 3: WHOIS fallback — when RDAP fails
// Phase 4: SSL — only when inconclusive
func (s *Scanner) check(domain string, workerID int) Result {
	parts := strings.SplitN(domain, ".", 2)
	if len(parts) != 2 {
		return Result{Domain: domain, Error: true, Timestamp: time.Now()}
	}
	tld := parts[1]

	var signatures []string

	// Phase 1: DNS check (round-robin across resolvers)
	resolver := s.resolvers[workerID%len(s.resolvers)]
	dnsResult, dnsSigs := s.checkDNS(domain, resolver)
	signatures = append(signatures, dnsSigs...)

	if dnsResult == "taken" {
		return Result{Domain: domain, TLD: tld, Available: false, Method: "dns", Signatures: signatures, Timestamp: time.Now()}
	}

	if dnsResult == "error" {
		// DNS failed — try RDAP directly
		r := s.checkRDAP(domain, tld)
		if r.Available || !r.Error {
			r.Signatures = signatures
			if !r.Available {
				r.Signatures = append(r.Signatures, "RDAP")
			}
			return r
		}
		// RDAP also failed — try WHOIS
		whoisResult := checkWHOIS(domain)
		if whoisResult == "taken" {
			signatures = append(signatures, "WHOIS")
			return Result{Domain: domain, TLD: tld, Available: false, Method: "whois", Signatures: signatures, Timestamp: time.Now()}
		}
		if whoisResult == "available" {
			return Result{Domain: domain, TLD: tld, Available: true, Method: "whois", Signatures: signatures, Timestamp: time.Now()}
		}
		return Result{Domain: domain, TLD: tld, Error: true, Method: "dns", Signatures: signatures, Timestamp: time.Now()}
	}

	// Phase 2: DNS says NXDOMAIN — confirm with RDAP
	baseURL := GetRDAPEndpoint(tld)
	if baseURL == "" {
		// No RDAP — try WHOIS as fallback
		whoisResult := checkWHOIS(domain)
		if whoisResult == "taken" {
			signatures = append(signatures, "WHOIS")
			return Result{Domain: domain, TLD: tld, Available: false, Method: "whois", Signatures: signatures, Timestamp: time.Now()}
		}
		return Result{Domain: domain, TLD: tld, Available: true, Method: "dns", Signatures: signatures, Timestamp: time.Now()}
	}

	rdapResult := s.checkRDAPDirect(domain, baseURL)
	switch rdapResult {
	case "available":
		return Result{Domain: domain, TLD: tld, Available: true, Method: "dns+rdap", Signatures: signatures, Timestamp: time.Now()}
	case "taken":
		signatures = append(signatures, "RDAP")
		return Result{Domain: domain, TLD: tld, Available: false, Method: "dns+rdap", Signatures: signatures, Timestamp: time.Now()}
	default:
		// RDAP error — use WHOIS as fallback
		whoisResult := checkWHOIS(domain)
		switch whoisResult {
		case "available":
			return Result{Domain: domain, TLD: tld, Available: true, Method: "dns+whois", Signatures: signatures, Timestamp: time.Now()}
		case "taken":
			signatures = append(signatures, "WHOIS")
			return Result{Domain: domain, TLD: tld, Available: false, Method: "dns+whois", Signatures: signatures, Timestamp: time.Now()}
		default:
			// All inconclusive — try SSL as last resort
			if checkSSL(domain) {
				signatures = append(signatures, "SSL")
				return Result{Domain: domain, TLD: tld, Available: false, Method: "ssl", Signatures: signatures, Timestamp: time.Now()}
			}
			// Trust DNS NXDOMAIN
			return Result{Domain: domain, TLD: tld, Available: true, Method: "dns", Signatures: signatures, Timestamp: time.Now()}
		}
	}
}

// checkDNS queries NS, A, and MX records. Returns status and signatures of methods that confirmed "taken".
func (s *Scanner) checkDNS(domain string, resolver *net.Resolver) (string, []string) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var sigs []string

	// Check NS records first (most authoritative)
	_, err := resolver.LookupNS(ctx, domain)
	if err == nil {
		sigs = append(sigs, "DNS_NS")
		return "taken", sigs
	}

	// Check if NXDOMAIN
	if dnsErr, ok := err.(*net.DNSError); ok {
		if dnsErr.IsNotFound {
			return "available", nil
		}
		if dnsErr.IsTemporary {
			return "error", nil
		}
	}

	// Fallback: try A record
	ctx2, cancel2 := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel2()

	ips, err := resolver.LookupHost(ctx2, domain)
	if err == nil && len(ips) > 0 {
		sigs = append(sigs, "DNS_A")
		return "taken", sigs
	}

	if dnsErr, ok := err.(*net.DNSError); ok && dnsErr.IsNotFound {
		return "available", nil
	}

	// Fallback: try MX record
	ctx3, cancel3 := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel3()

	mxRecords, err := resolver.LookupMX(ctx3, domain)
	if err == nil && len(mxRecords) > 0 {
		sigs = append(sigs, "DNS_MX")
		return "taken", sigs
	}

	if dnsErr, ok := err.(*net.DNSError); ok && dnsErr.IsNotFound {
		return "available", nil
	}

	return "error", nil
}

// checkRDAP does a full RDAP check (used when DNS fails entirely).
func (s *Scanner) checkRDAP(domain, tld string) Result {
	baseURL := GetRDAPEndpoint(tld)
	if baseURL == "" {
		return Result{Domain: domain, TLD: tld, Error: true, Method: "rdap", Timestamp: time.Now()}
	}

	result := s.checkRDAPDirect(domain, baseURL)
	switch result {
	case "available":
		return Result{Domain: domain, TLD: tld, Available: true, Method: "rdap", Timestamp: time.Now()}
	case "taken":
		return Result{Domain: domain, TLD: tld, Available: false, Method: "rdap", Timestamp: time.Now()}
	default:
		return Result{Domain: domain, TLD: tld, Error: true, Method: "rdap", Timestamp: time.Now()}
	}
}

// checkRDAPDirect makes the HTTP request. Returns "available", "taken", or "error".
func (s *Scanner) checkRDAPDirect(domain, baseURL string) string {
	url := baseURL + domain

	var resp *http.Response
	var err error

	for attempt := 0; attempt < 2; attempt++ {
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Accept", "application/rdap+json")

		resp, err = s.httpClient.Do(req)
		if err == nil {
			break
		}
		time.Sleep(time.Duration(200*(1+attempt)) * time.Millisecond)
	}

	if err != nil {
		return "error"
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 404, 400:
		return "available"
	case 200:
		return "taken"
	case 429:
		time.Sleep(1 * time.Second)
		return "error"
	default:
		return "error"
	}
}

// GenerateAll creates domains for given TLDs, length, and pattern.
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
