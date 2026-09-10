package trademark

import (
	"fmt"
	"net/url"
)

// SearchURLs holds links to trademark search portals for a given name.
type SearchURLs struct {
	INPI  string `json:"inpi,omitempty"`
	USPTO string `json:"uspto,omitempty"`
	EUIPO string `json:"euipo,omitempty"`
}

// For returns trademark search URLs for the given base name.
func For(name string) SearchURLs {
	q := url.QueryEscape(name)
	return SearchURLs{
		INPI:  fmt.Sprintf("https://busca.inpi.gov.br/pePI/servlet/MarcasServlet?action=searchBasico&tipo=EM&nrProcesso=&nmMarca=%s", q),
		USPTO: fmt.Sprintf("https://tmsearch.uspto.gov/search/search-information?searchInput=%s&searchOption=All+Design+Search+Codes", q),
		EUIPO: fmt.Sprintf("https://euipo.europa.eu/eSearch/#advanced/trademarks/wordMark=%s", q),
	}
}
