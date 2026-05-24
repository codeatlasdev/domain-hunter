package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/codeatlasdev/domain-hunter/internal/registry"
)

type tldEntry struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	HasRDAP bool   `json:"has_rdap"`
}

func HandleTLDs(w http.ResponseWriter, r *http.Request) {
	tlds, err := registry.GetCachedTLDs()
	if err != nil {
		http.Error(w, `{"error":"failed to fetch TLDs"}`, http.StatusInternalServerError)
		return
	}

	rdapCount := 0
	entries := make([]tldEntry, len(tlds))
	for i, t := range tlds {
		hasRDAP := t.RDAPUrl != ""
		if hasRDAP {
			rdapCount++
		}
		entries[i] = tldEntry{Name: t.Name, Type: t.Type, HasRDAP: hasRDAP}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"total":          len(tlds),
		"rdap_supported": rdapCount,
		"tlds":           entries,
	})
}
