package registry

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type TLDInfo struct {
	Name    string `json:"name"`
	RDAPUrl string `json:"rdap_url"`
	Type    string `json:"type"`
}

type rdapBootstrap struct {
	Services [][][]string `json:"services"`
}

type cache struct {
	FetchedAt time.Time `json:"fetched_at"`
	TLDs      []TLDInfo `json:"tlds"`
}

type rdapCache struct {
	FetchedAt time.Time         `json:"fetched_at"`
	Endpoints map[string]string `json:"endpoints"`
}

const cacheTTL = 24 * time.Hour

func cacheDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".domh")
	return dir, os.MkdirAll(dir, 0o755)
}

func FetchAllTLDs() ([]TLDInfo, error) {
	resp, err := http.Get("https://data.iana.org/TLD/tlds-alpha-by-domain.txt")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	rdapMap, _ := FetchRDAPBootstrap()

	// Known country-code TLDs (2-letter)
	var tlds []TLDInfo
	for _, line := range strings.Split(string(body), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		name := strings.ToLower(line)
		t := "generic"
		if len(name) == 2 {
			t = "country-code"
		}
		rdapURL := ""
		if rdapMap != nil {
			rdapURL = rdapMap[name]
		}
		tlds = append(tlds, TLDInfo{Name: name, RDAPUrl: rdapURL, Type: t})
	}

	// Save cache
	dir, err := cacheDir()
	if err == nil {
		data, _ := json.Marshal(cache{FetchedAt: time.Now(), TLDs: tlds})
		_ = os.WriteFile(filepath.Join(dir, "tlds.json"), data, 0o644)
	}

	return tlds, nil
}

func FetchRDAPBootstrap() (map[string]string, error) {
	resp, err := http.Get("https://data.iana.org/rdap/dns.json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var bs rdapBootstrap
	if err := json.NewDecoder(resp.Body).Decode(&bs); err != nil {
		return nil, err
	}

	endpoints := make(map[string]string)
	for _, svc := range bs.Services {
		if len(svc) < 2 || len(svc[1]) == 0 {
			continue
		}
		url := svc[1][0]
		for _, tld := range svc[0] {
			endpoints[strings.ToLower(tld)] = url
		}
	}

	// Save cache
	dir, err := cacheDir()
	if err == nil {
		data, _ := json.Marshal(rdapCache{FetchedAt: time.Now(), Endpoints: endpoints})
		_ = os.WriteFile(filepath.Join(dir, "rdap.json"), data, 0o644)
	}

	return endpoints, nil
}

func GetCachedTLDs() ([]TLDInfo, error) {
	dir, err := cacheDir()
	if err != nil {
		return FetchAllTLDs()
	}

	data, err := os.ReadFile(filepath.Join(dir, "tlds.json"))
	if err != nil {
		return FetchAllTLDs()
	}

	var c cache
	if err := json.Unmarshal(data, &c); err != nil {
		return FetchAllTLDs()
	}

	if time.Since(c.FetchedAt) > cacheTTL {
		return FetchAllTLDs()
	}

	return c.TLDs, nil
}

func GetRDAPEndpoint(tld string) string {
	tld = strings.ToLower(tld)

	// Try cached rdap.json first
	dir, err := cacheDir()
	if err == nil {
		data, err := os.ReadFile(filepath.Join(dir, "rdap.json"))
		if err == nil {
			var c rdapCache
			if json.Unmarshal(data, &c) == nil && time.Since(c.FetchedAt) <= cacheTTL {
				if url, ok := c.Endpoints[tld]; ok {
					return url + "domain/"
				}
				return ""
			}
		}
	}

	// Fetch fresh
	endpoints, err := FetchRDAPBootstrap()
	if err != nil {
		return ""
	}
	if url, ok := endpoints[tld]; ok {
		return url + "domain/"
	}
	return ""
}

func RefreshCache() error {
	_, err := FetchAllTLDs()
	if err != nil {
		return fmt.Errorf("fetching TLDs: %w", err)
	}
	return nil
}
