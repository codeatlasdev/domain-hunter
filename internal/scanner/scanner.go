package scanner

import (
	"crypto/tls"
	"math/rand"
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
	clients map[string]*http.Client
	workers int
	stats   Stats
	mu      sync.Mutex
	results []Result

	OnResult func(Result)
	Done     chan struct{}
}

func New(workers int) *Scanner {
	clients := make(map[string]*http.Client)
	transport := &http.Transport{
		MaxIdleConns:        300,
		MaxIdleConnsPerHost: 100,
		MaxConnsPerHost:     100,
		IdleConnTimeout:     90 * time.Second,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: false},
		DisableKeepAlives:   false,
	}
	for tld := range Providers {
		clients[tld] = &http.Client{
			Timeout:   10 * time.Second,
			Transport: transport,
		}
	}
	return &Scanner{
		clients: clients,
		workers: workers,
		Done:    make(chan struct{}),
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

	// Shuffle to distribute load across TLDs
	shuffled := make([]string, len(domains))
	copy(shuffled, domains)
	rand.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })

	work := make(chan string, s.workers*2)
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
	tld := parts[1]

	baseURL, ok := Providers[tld]
	if !ok {
		return Result{Domain: domain, TLD: tld, Error: true, Timestamp: time.Now()}
	}

	client := s.clients[tld]
	url := baseURL + domain

	var resp *http.Response
	var err error

	// Retry up to 2 times with backoff
	for attempt := 0; attempt < 3; attempt++ {
		resp, err = client.Get(url)
		if err == nil {
			break
		}
		if attempt < 2 {
			time.Sleep(time.Duration(100*(attempt+1)) * time.Millisecond)
		}
	}

	if err != nil {
		return Result{Domain: domain, TLD: tld, Error: true, Timestamp: time.Now()}
	}
	defer resp.Body.Close()

	// 429 = rate limited, treat as error
	if resp.StatusCode == 429 {
		time.Sleep(500 * time.Millisecond)
		return Result{Domain: domain, TLD: tld, Error: true, Timestamp: time.Now()}
	}

	// 404 or 400 = not found = available
	available := resp.StatusCode == 404 || resp.StatusCode == 400

	return Result{Domain: domain, TLD: tld, Available: available, Timestamp: time.Now()}
}
