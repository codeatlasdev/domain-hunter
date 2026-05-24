package scanner

// RDAP providers by TLD — sourced from IANA bootstrap (data.iana.org/rdap/dns.json).
// Protocol: HTTP GET → 200 = registered, 404 = available.
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
