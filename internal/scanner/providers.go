package scanner

// RDAP providers by TLD. 404 = available, 200 = taken.
var Providers = map[string]string{
	"com": "https://rdap.verisign.com/com/v1/domain/",
	"net": "https://rdap.verisign.com/net/v1/domain/",
	"org": "https://rdap.org/domain/",
	"dev": "https://rdap.nic.google/domain/",
	"app": "https://rdap.nic.google/domain/",
	"io":  "https://rdap.nic.io/domain/",
	"co":  "https://rdap.nic.co/domain/",
	"xyz": "https://rdap.nic.xyz/domain/",
}

var SupportedTLDs = []string{"com", "net", "org", "dev", "app", "io", "co", "xyz"}
