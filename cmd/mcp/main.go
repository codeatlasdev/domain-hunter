package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/codeatlasdev/domain-hunter/internal/presets"
	"github.com/codeatlasdev/domain-hunter/internal/pricing"
	"github.com/codeatlasdev/domain-hunter/internal/scanner"
)

func main() {
	s := server.NewMCPServer(
		"domh",
		"1.5.0",
		server.WithToolCapabilities(false),
	)

	s.AddTool(mcp.NewTool("check_domain",
		mcp.WithDescription("Check if a domain name is available for registration"),
		mcp.WithString("domain", mcp.Required(), mcp.Description("Full domain name to check (e.g. coolname.com)")),
	), checkDomainHandler)

	s.AddTool(mcp.NewTool("check_domains",
		mcp.WithDescription("Check multiple domain names for availability"),
		mcp.WithString("domains", mcp.Required(), mcp.Description("Comma-separated list of domains to check")),
	), checkDomainsHandler)

	s.AddTool(mcp.NewTool("check_with_preset",
		mcp.WithDescription("Check a name across a curated set of TLDs"),
		mcp.WithString("name", mcp.Required(), mcp.Description("Base name to check (without TLD)")),
		mcp.WithString("preset", mcp.Required(), mcp.Description("Preset name: startup, tech, creative, ecommerce, finance, popular, classic, enterprise, web, trendy, country, brazil")),
	), checkWithPresetHandler)

	s.AddTool(mcp.NewTool("generate_names",
		mcp.WithDescription("Generate pronounceable domain names by pattern"),
		mcp.WithNumber("length", mcp.Required(), mcp.Description("Name length: 3, 4, or 5")),
		mcp.WithString("pattern", mcp.Description("Pattern: CVC, VCV, CVCV, CVCVC, ALL. Default: ALL")),
		mcp.WithString("tld", mcp.Description("TLD to append. Default: com")),
	), generateNamesHandler)

	s.AddTool(mcp.NewTool("get_prices",
		mcp.WithDescription("Get registrar prices and buy links for a domain"),
		mcp.WithString("domain", mcp.Required(), mcp.Description("Domain to get prices for (e.g. coolname.com)")),
	), getPricesHandler)

	s.AddTool(mcp.NewTool("list_presets",
		mcp.WithDescription("List all available TLD presets with their TLDs"),
	), listPresetsHandler)

	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

func checkDomainHandler(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	domain, err := req.RequireString("domain")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	domain = strings.TrimSpace(strings.ToLower(domain))
	if !strings.Contains(domain, ".") {
		return mcp.NewToolResultError("domain must include TLD (e.g. coolname.com)"), nil
	}

	r := scanner.CheckSingle(domain)
	b, _ := json.Marshal(r)
	return mcp.NewToolResultText(string(b)), nil
}

func checkDomainsHandler(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	raw, err := req.RequireString("domains")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var domains []string
	for _, d := range strings.Split(raw, ",") {
		d = strings.TrimSpace(strings.ToLower(d))
		if d != "" && strings.Contains(d, ".") {
			domains = append(domains, d)
		}
	}
	if len(domains) == 0 {
		return mcp.NewToolResultError("no valid domains provided"), nil
	}

	results := scanner.CheckMultiple(domains)
	b, _ := json.Marshal(results)
	return mcp.NewToolResultText(string(b)), nil
}

func checkWithPresetHandler(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, err := req.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	presetName, err := req.RequireString("preset")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	name = strings.TrimSpace(strings.ToLower(name))
	tlds, ok := presets.Get(presetName)
	if !ok {
		return mcp.NewToolResultError(fmt.Sprintf("unknown preset: %s", presetName)), nil
	}

	var domains []string
	for _, tld := range tlds {
		domains = append(domains, fmt.Sprintf("%s.%s", name, tld))
	}

	results := scanner.CheckMultiple(domains)
	b, _ := json.Marshal(results)
	return mcp.NewToolResultText(string(b)), nil
}

func generateNamesHandler(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	length := req.GetInt("length", 3)
	pattern := req.GetString("pattern", "ALL")
	tld := req.GetString("tld", "com")

	if length < 3 || length > 5 {
		return mcp.NewToolResultError("length must be 3, 4, or 5"), nil
	}

	names := scanner.Generate(length, strings.ToUpper(pattern))
	var domains []string
	for _, n := range names {
		domains = append(domains, fmt.Sprintf("%s.%s", n, tld))
	}

	type result struct {
		Count   int      `json:"count"`
		Domains []string `json:"domains"`
	}
	// Limit output to first 100 for readability
	out := domains
	if len(out) > 100 {
		out = out[:100]
	}
	b, _ := json.Marshal(result{Count: len(domains), Domains: out})
	return mcp.NewToolResultText(string(b)), nil
}

func getPricesHandler(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	domain, err := req.RequireString("domain")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	domain = strings.TrimSpace(strings.ToLower(domain))

	pr := pricing.GetPrices(domain)
	b, _ := json.Marshal(pr)
	return mcp.NewToolResultText(string(b)), nil
}

func listPresetsHandler(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	b, _ := json.Marshal(presets.List())
	return mcp.NewToolResultText(string(b)), nil
}
