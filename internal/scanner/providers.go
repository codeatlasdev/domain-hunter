package scanner

import "github.com/codeatlasdev/domain-hunter/internal/registry"

// Providers is kept for backward compatibility — populated from registry cache.
var Providers = map[string]string{
	"com": "https://rdap.verisign.com/com/v1/domain/",
	"net": "https://rdap.verisign.com/net/v1/domain/",
	"org": "https://rdap.publicinterestregistry.org/rdap/domain/",
	"dev": "https://pubapi.registry.google/rdap/domain/",
	"app": "https://pubapi.registry.google/rdap/domain/",
	"io":  "https://rdap.nic.io/domain/",
	"co":  "https://rdap.nic.co/domain/",
	"xyz": "https://rdap.centralnic.com/xyz/domain/",
}

var SupportedTLDs = []string{"com", "net", "org", "dev", "app", "io", "co", "xyz"}

// GetRDAPEndpoint resolves the RDAP endpoint for any TLD dynamically.
// Falls back to the hardcoded Providers map if registry has no data.
func GetRDAPEndpoint(tld string) string {
	if url, ok := Providers[tld]; ok {
		return url
	}
	return registry.GetRDAPEndpoint(tld)
}

// InitSupportedTLDs populates SupportedTLDs from the registry cache.
func InitSupportedTLDs() {
	tlds, err := registry.GetCachedTLDs()
	if err != nil || len(tlds) == 0 {
		return
	}
	all := make([]string, 0, len(tlds))
	for _, t := range tlds {
		all = append(all, t.Name)
	}
	SupportedTLDs = all
}
