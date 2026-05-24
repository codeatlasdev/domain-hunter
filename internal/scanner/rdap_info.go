package scanner

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
)

type rdapResponse struct {
	Entities []rdapEntity `json:"entities"`
	Events   []rdapEvent  `json:"events"`
	Status   []string     `json:"status"`
}

type rdapEntity struct {
	Roles      []string     `json:"roles"`
	VCardArray []any        `json:"vcardArray"`
	Entities   []rdapEntity `json:"entities"`
}

type rdapEvent struct {
	EventAction string `json:"eventAction"`
	EventDate   string `json:"eventDate"`
}

// FetchDomainInfo queries RDAP for registrar/dates info on a taken domain.
func FetchDomainInfo(domain, tld string) *DomainInfo {
	baseURL := GetRDAPEndpoint(tld)
	if baseURL == "" {
		return nil
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", baseURL+domain, nil)
	req.Header.Set("Accept", "application/rdap+json")

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	var rdap rdapResponse
	if err := json.Unmarshal(body, &rdap); err != nil {
		return nil
	}

	info := &DomainInfo{Status: rdap.Status}

	for _, e := range rdap.Events {
		date := formatRDAPDate(e.EventDate)
		switch e.EventAction {
		case "registration":
			info.CreatedDate = date
		case "expiration":
			info.ExpiryDate = date
		}
	}

	for _, ent := range rdap.Entities {
		for _, role := range ent.Roles {
			if role == "registrar" {
				info.Registrar = extractVCardName(ent)
				break
			}
		}
	}

	return info
}

func formatRDAPDate(s string) string {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.Format("2006-01-02")
	}
	if len(s) >= 10 {
		return s[:10]
	}
	return s
}

func extractVCardName(ent rdapEntity) string {
	if len(ent.VCardArray) >= 2 {
		if props, ok := ent.VCardArray[1].([]any); ok {
			for _, prop := range props {
				if arr, ok := prop.([]any); ok && len(arr) >= 4 {
					if name, ok := arr[0].(string); ok && name == "fn" {
						if val, ok := arr[3].(string); ok {
							return strings.TrimSpace(val)
						}
					}
				}
			}
		}
	}
	return ""
}
