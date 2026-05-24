package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/codeatlasdev/domain-hunter/internal/pricing"
	"github.com/codeatlasdev/domain-hunter/internal/scanner"
)

type checkRequest struct {
	Keyword string   `json:"keyword"`
	TLDs    []string `json:"tlds"`
	Prefix  string   `json:"prefix"`
	Suffix  string   `json:"suffix"`
}

type domainResult struct {
	Domain    string               `json:"domain"`
	Available bool                 `json:"available"`
	Method    string               `json:"method"`
	Pricing   *pricing.PriceResult `json:"pricing,omitempty"`
}

type checkResponse struct {
	Results []domainResult `json:"results"`
	Stats   checkStats     `json:"stats"`
}

type checkStats struct {
	Total     int   `json:"total"`
	Available int   `json:"available"`
	Taken     int   `json:"taken"`
	ElapsedMs int64 `json:"elapsed_ms"`
}

func HandleCheck(w http.ResponseWriter, r *http.Request) {
	var req checkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}

	if req.Keyword == "" {
		http.Error(w, `{"error":"keyword required"}`, http.StatusBadRequest)
		return
	}
	if len(req.TLDs) == 0 {
		req.TLDs = []string{"com"}
	}
	if len(req.TLDs) > 50 {
		http.Error(w, `{"error":"max 50 TLDs"}`, http.StatusBadRequest)
		return
	}

	domains := buildDomains(req.Keyword, req.TLDs, req.Prefix, req.Suffix)
	start := time.Now()
	results := scanner.CheckMultiple(domains)

	var out []domainResult
	available := 0
	for _, res := range results {
		dr := domainResult{Domain: res.Domain, Available: res.Available, Method: res.Method}
		if res.Available {
			available++
			dr.Pricing = res.Pricing
		}
		out = append(out, dr)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(checkResponse{
		Results: out,
		Stats: checkStats{
			Total:     len(domains),
			Available: available,
			Taken:     len(domains) - available,
			ElapsedMs: time.Since(start).Milliseconds(),
		},
	})
}

func HandleCheckStream(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("keyword")
	if keyword == "" {
		http.Error(w, `{"error":"keyword required"}`, http.StatusBadRequest)
		return
	}
	tldsParam := r.URL.Query().Get("tlds")
	if tldsParam == "" {
		tldsParam = "com"
	}
	tlds := strings.Split(tldsParam, ",")
	if len(tlds) > 50 {
		http.Error(w, `{"error":"max 50 TLDs"}`, http.StatusBadRequest)
		return
	}
	prefix := r.URL.Query().Get("prefix")
	suffix := r.URL.Query().Get("suffix")

	domains := buildDomains(keyword, tlds, prefix, suffix)

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, `{"error":"streaming not supported"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	start := time.Now()
	available := 0
	ctx := r.Context()

	// Channel serializes writes from scanner goroutines to this single goroutine.
	// Closed by the sender goroutine after scanner finishes.
	results := make(chan domainResult, 64)

	s := scanner.New(10)
	s.OnResult = func(res scanner.Result) {
		dr := domainResult{Domain: res.Domain, Available: res.Available, Method: res.Method}
		if res.Available {
			dr.Pricing = res.Pricing
		}
		select {
		case results <- dr:
		case <-ctx.Done():
		}
	}

	go func() {
		s.Run(domains)
		<-s.Done
		close(results)
	}()

	for {
		select {
		case dr, ok := <-results:
			if !ok {
				goto done
			}
			if dr.Available {
				available++
			}
			data, _ := json.Marshal(dr)
			fmt.Fprintf(w, "event: result\ndata: %s\n\n", data)
			flusher.Flush()
		case <-ctx.Done():
			return
		}
	}
done:
	stats := checkStats{
		Total:     len(domains),
		Available: available,
		Taken:     len(domains) - available,
		ElapsedMs: time.Since(start).Milliseconds(),
	}
	data, _ := json.Marshal(stats)
	fmt.Fprintf(w, "event: done\ndata: %s\n\n", data)
	flusher.Flush()
}

func buildDomains(keyword string, tlds []string, prefix, suffix string) []string {
	keyword = strings.ToLower(strings.TrimSpace(keyword))
	var names []string
	if prefix == "" && suffix == "" {
		names = []string{keyword}
	} else {
		prefixes := splitTrim(prefix)
		suffixes := splitTrim(suffix)
		if len(prefixes) == 0 {
			prefixes = []string{""}
		}
		if len(suffixes) == 0 {
			suffixes = []string{""}
		}
		for _, p := range prefixes {
			for _, s := range suffixes {
				names = append(names, p+keyword+s)
			}
		}
	}

	var domains []string
	for _, tld := range tlds {
		tld = strings.ToLower(strings.TrimSpace(tld))
		for _, name := range names {
			domains = append(domains, name+"."+tld)
		}
	}
	return domains
}

func splitTrim(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
