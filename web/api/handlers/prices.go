package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/codeatlasdev/domain-hunter/internal/pricing"
)

func HandlePrices(w http.ResponseWriter, r *http.Request) {
	// Route: /api/prices/{domain}
	domain := strings.TrimPrefix(r.URL.Path, "/api/prices/")
	if domain == "" || !strings.Contains(domain, ".") {
		http.Error(w, `{"error":"domain required (e.g. /api/prices/cool.com)"}`, http.StatusBadRequest)
		return
	}

	result := pricing.GetPrices(strings.ToLower(domain))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
