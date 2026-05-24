package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/codeatlasdev/domain-hunter/internal/presets"
)

func HandlePresets(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"presets": presets.List(),
	})
}
