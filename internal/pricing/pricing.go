package pricing

import (
	"fmt"
	"sort"
	"strings"
)

type Price struct {
	Registrar     string  `json:"registrar"`
	RegisterPrice float64 `json:"register_price"`
	RenewPrice    float64 `json:"renew_price"`
	TransferPrice float64 `json:"transfer_price"`
	Currency      string  `json:"currency"`
	BuyURL        string  `json:"buy_url"`
}

type PriceResult struct {
	Domain   string  `json:"domain"`
	TLD      string  `json:"tld"`
	Prices   []Price `json:"prices"`
	Cheapest *Price  `json:"cheapest"`
}

var buyLinks = map[string]func(string) string{
	"Namecheap":   func(d string) string { return fmt.Sprintf("https://www.namecheap.com/domains/registration/results/?domain=%s", d) },
	"Porkbun":     func(d string) string { return fmt.Sprintf("https://porkbun.com/checkout/search?q=%s", d) },
	"Cloudflare":  func(d string) string { return "https://www.cloudflare.com/products/registrar/" },
	"GoDaddy":     func(d string) string { return fmt.Sprintf("https://www.godaddy.com/domainsearch/find?domainToCheck=%s", d) },
	"Google":      func(d string) string { return fmt.Sprintf("https://domains.google.com/registrar/search?searchTerm=%s", d) },
	"Dynadot":     func(d string) string { return fmt.Sprintf("https://www.dynadot.com/domain/search?domain=%s", d) },
	"NameSilo":    func(d string) string { return fmt.Sprintf("https://www.namesilo.com/domain/search-domains?query=%s", d) },
	"Spaceship":   func(d string) string { return fmt.Sprintf("https://www.spaceship.com/domain/%s", d) },
	"Hostinger":   func(d string) string { return fmt.Sprintf("https://www.hostinger.com/domain-name-search?query=%s", d) },
	"IONOS":       func(d string) string { return fmt.Sprintf("https://www.ionos.com/domains/%s", d) },
	"OVH":         func(d string) string { return "https://www.ovhcloud.com/en/domains/" },
	"Gandi":       func(d string) string { return fmt.Sprintf("https://www.gandi.net/en/domain/search?query=%s", d) },
	"Hover":       func(d string) string { return fmt.Sprintf("https://www.hover.com/domains/results?q=%s", d) },
	"Epik":        func(d string) string { return fmt.Sprintf("https://www.epik.com/domain/%s", d) },
	"Name.com":    func(d string) string { return fmt.Sprintf("https://www.name.com/domain/search/%s", d) },
	"Registro.br": func(d string) string {
		name := strings.Split(d, ".")[0]
		return fmt.Sprintf("https://registro.br/dominio/novo/?q=%s", name)
	},
	"101domain":   func(d string) string { return fmt.Sprintf("https://www.101domain.com/search/?q=%s", d) },
	"INWX":        func(d string) string { return fmt.Sprintf("https://www.inwx.com/en/domain/check#search=%s", d) },
	"Dreamhost":   func(d string) string { return fmt.Sprintf("https://www.dreamhost.com/domains/?domain=%s", d) },
}

var knownPrices = map[string]map[string]float64{
	"com":     {"Cloudflare": 8.57, "Porkbun": 8.88, "Namecheap": 9.98, "NameSilo": 9.95, "Spaceship": 7.48, "GoDaddy": 11.99, "Google": 12.00, "Hostinger": 9.99},
	"dev":     {"Cloudflare": 10.11, "Porkbun": 10.58, "Namecheap": 12.98, "Google": 12.00, "GoDaddy": 14.99},
	"io":      {"Cloudflare": 33.98, "Porkbun": 28.88, "Namecheap": 32.98, "GoDaddy": 39.99, "Spaceship": 29.98},
	"app":     {"Cloudflare": 14.00, "Porkbun": 13.88, "Namecheap": 14.98, "Google": 14.00, "GoDaddy": 16.99},
	"net":     {"Cloudflare": 10.11, "Porkbun": 10.28, "Namecheap": 12.98, "NameSilo": 11.79, "GoDaddy": 14.99},
	"org":     {"Cloudflare": 9.93, "Porkbun": 9.88, "Namecheap": 11.98, "NameSilo": 10.79, "GoDaddy": 12.99},
	"co":      {"Cloudflare": 11.50, "Porkbun": 11.18, "Namecheap": 12.98, "GoDaddy": 14.99},
	"xyz":     {"Cloudflare": 9.94, "Porkbun": 9.00, "Namecheap": 10.98, "Spaceship": 2.48, "GoDaddy": 12.99},
	"ai":      {"Porkbun": 26.88, "Namecheap": 74.95, "GoDaddy": 79.99, "101domain": 58.00},
	// .br TLDs — R$40/year, only NIC.br registers them (no third-party registrars)
	"com.br":  {"Registro.br": 40.00},
	"net.br":  {"Registro.br": 40.00},
	"org.br":  {"Registro.br": 40.00},
	"app.br":  {"Registro.br": 40.00},
	"dev.br":  {"Registro.br": 40.00},
	"tmp.br":  {"Registro.br": 40.00},
	"blog.br": {"Registro.br": 40.00},
	"wiki.br": {"Registro.br": 40.00},
	"nom.br":  {"Registro.br": 40.00},
	"pro.br":  {"Registro.br": 40.00},
	"mus.br":  {"Registro.br": 40.00},
	"art.br":  {"Registro.br": 40.00},
	"eco.br":  {"Registro.br": 40.00},
	"rec.br":  {"Registro.br": 40.00},
	"srv.br":  {"Registro.br": 40.00},
	"adv.br":  {"Registro.br": 40.00},
	"eng.br":  {"Registro.br": 40.00},
	"med.br":  {"Registro.br": 40.00},
	"psi.br":  {"Registro.br": 40.00},
	"vet.br":  {"Registro.br": 40.00},
	"arq.br":  {"Registro.br": 40.00},
	"jor.br":  {"Registro.br": 40.00},
}

// tldCurrency maps TLDs that price in a non-USD currency.
var tldCurrency = map[string]string{
	"com.br":  "BRL",
	"net.br":  "BRL",
	"org.br":  "BRL",
	"app.br":  "BRL",
	"dev.br":  "BRL",
	"tmp.br":  "BRL",
	"blog.br": "BRL",
	"wiki.br": "BRL",
	"nom.br":  "BRL",
	"pro.br":  "BRL",
	"mus.br":  "BRL",
	"art.br":  "BRL",
	"eco.br":  "BRL",
	"rec.br":  "BRL",
	"srv.br":  "BRL",
	"adv.br":  "BRL",
	"eng.br":  "BRL",
	"med.br":  "BRL",
	"psi.br":  "BRL",
	"vet.br":  "BRL",
	"arq.br":  "BRL",
	"jor.br":  "BRL",
}

func GetPrices(domain string) PriceResult {
	parts := strings.SplitN(domain, ".", 2)
	if len(parts) != 2 {
		return PriceResult{Domain: domain}
	}
	tld := parts[1]

	currency := "USD"
	if c, ok := tldCurrency[tld]; ok {
		currency = c
	}

	var prices []Price

	if tldPrices, ok := knownPrices[tld]; ok {
		for registrar, price := range tldPrices {
			buyURL := ""
			if linkFn, ok := buyLinks[registrar]; ok {
				buyURL = linkFn(domain)
			}
			prices = append(prices, Price{
				Registrar:     registrar,
				RegisterPrice: price,
				Currency:      currency,
				BuyURL:        buyURL,
			})
		}
	}

	sort.Slice(prices, func(i, j int) bool {
		return prices[i].RegisterPrice < prices[j].RegisterPrice
	})

	result := PriceResult{Domain: domain, TLD: tld, Prices: prices}
	if len(prices) > 0 {
		result.Cheapest = &prices[0]
	}
	return result
}
