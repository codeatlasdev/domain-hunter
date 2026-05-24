package scanner

import (
	"strings"

	"github.com/likexian/whois"
)

var whoisServers = []string{
	"",
	"whois.verisign-grs.com:43",
	"whois.internic.net:43",
}

var availableIndicators = []string{
	"no match for", "not found", "no data found",
	"no entries found", "domain not found", "no object found",
	"status: free", "status: available", "available for registration",
	"not registered", "no match",
}

var registeredIndicators = []string{
	"registrar:", "creation date:", "name server:",
	"expiration date:", "updated date:", "nserver:",
}

func checkWHOIS(domain string) string {
	for _, server := range whoisServers {
		var result string
		var err error
		if server == "" {
			result, err = whois.Whois(domain)
		} else {
			result, err = whois.Whois(domain, server)
		}
		if err != nil || result == "" {
			continue
		}
		lower := strings.ToLower(result)
		for _, ind := range availableIndicators {
			if strings.Contains(lower, ind) {
				return "available"
			}
		}
		for _, ind := range registeredIndicators {
			if strings.Contains(lower, ind) {
				return "taken"
			}
		}
	}
	return "unknown"
}
