package presets

var Presets = map[string][]string{
	"startup":    {"com", "org", "io", "ai", "tech", "app", "dev", "xyz"},
	"popular":    {"com", "net", "org", "io", "ai", "app", "dev", "tech", "me", "co", "xyz"},
	"classic":    {"com", "net", "org", "info", "biz"},
	"enterprise": {"com", "org", "net", "info", "biz", "us"},
	"tech":       {"io", "ai", "app", "dev", "tech", "cloud", "software", "code", "systems"},
	"creative":   {"design", "art", "studio", "media", "photography", "gallery", "ink"},
	"ecommerce":  {"shop", "store", "market", "sale", "deals", "buy", "shopping"},
	"finance":    {"finance", "capital", "fund", "money", "investments", "bank", "pay"},
	"web":        {"web", "site", "website", "online", "blog", "page", "digital"},
	"trendy":     {"xyz", "online", "site", "top", "icu", "fun", "space", "life"},
	"country":    {"us", "uk", "de", "fr", "ca", "au", "br", "in", "nl", "jp"},
	"brazil":     {"com.br", "net.br", "org.br", "app.br", "dev.br"},
}

func Get(name string) ([]string, bool) {
	p, ok := Presets[name]
	return p, ok
}

func List() map[string][]string {
	return Presets
}
