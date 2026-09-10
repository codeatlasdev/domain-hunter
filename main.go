package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	selfupdate "github.com/creativeprojects/go-selfupdate"
	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"

	"github.com/codeatlasdev/domain-hunter/internal/export"
	"github.com/codeatlasdev/domain-hunter/internal/presets"
	"github.com/codeatlasdev/domain-hunter/internal/pricing"
	"github.com/codeatlasdev/domain-hunter/internal/registry"
	"github.com/codeatlasdev/domain-hunter/internal/scanner"
	"github.com/codeatlasdev/domain-hunter/internal/trademark"
	"github.com/codeatlasdev/domain-hunter/internal/tui"
	"github.com/codeatlasdev/domain-hunter/internal/wizard"
)

var (
	version = "dev"
	commit  = "none"
)

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#2563EB"))
	greenBold  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#10B981"))
	dimStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#64748B"))
	warnStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F59E0B"))
	redBold    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#EF4444"))
)

func main() {
	if len(os.Args) < 2 {
		runInteractive()
		return
	}

	switch os.Args[1] {
	case "scan":
		runCLI(os.Args[2:])
	case "check":
		runCheck(os.Args[2:])
	case "suggest":
		runSuggest(os.Args[2:])
	case "tlds":
		runTLDs(os.Args[2:])
	case "presets":
		runPresets()
	case "mcp":
		runMCP()
	case "update":
		runUpdate()
	case "version":
		fmt.Printf("domh %s (%s)\n", version, commit)
	case "help", "--help", "-h":
		printHelp()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", os.Args[1])
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println(titleStyle.Render("◆ domh") + " — bulk domain availability checker")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  domh                          Interactive wizard")
	fmt.Println("  domh scan [name] [flags]      Scan domains")
	fmt.Println("  domh check <file> [flags]     Dictionary mode")
	fmt.Println("  domh suggest [names] [flags]  Check name list, output JSON (agent-friendly)")
	fmt.Println("  domh tlds [flags]             List TLDs")
	fmt.Println("  domh presets                  List presets")
	fmt.Println("  domh mcp                      Start MCP server (stdio)")
	fmt.Println("  domh update                   Self-update")
	fmt.Println("  domh version                  Version info")
	fmt.Println()
	fmt.Println("Scan flags:")
	fmt.Println("  --tld          TLDs (comma-separated)       [default: com]")
	fmt.Println("  --preset       Use preset TLD set (startup, tech, etc)")
	fmt.Println("  --all          Check ALL 1437 TLDs")
	fmt.Println("  --length       Domain length (3-5)          [default: 3]")
	fmt.Println("  --pattern      CVC, VCV, CVCV, ALL          [default: ALL]")
	fmt.Println("  --prefix       Prefixes (comma-separated)")
	fmt.Println("  --suffix       Suffixes (comma-separated)")
	fmt.Println("  --workers      Concurrent workers           [default: 50]")
	fmt.Println("  --format       Export: txt,json,csv         [default: txt]")
	fmt.Println("  --regex, -r    Regex filter")
	fmt.Println("  --delay        Delay between queries (ms)   [default: 0]")
	fmt.Println("  --info         Show registrar info for taken domains")
	fmt.Println("  --show-registered  Save registered domains")
	fmt.Println("  --dry-run      Preview domains without checking")
	fmt.Println("  --yes, -y      Skip confirmations")
	fmt.Println("  --force        Skip performance warnings")
	fmt.Println("  --batch, --no-tui  Plain output (no TUI, CI/agent-friendly)")
	fmt.Println()
	fmt.Println("Suggest flags:")
	fmt.Println("  --tld          TLDs (comma-separated)       [default: com,io,app]")
	fmt.Println("  --preset       Use preset TLD set (saas, startup, brazil, global-saas, etc)")
	fmt.Println("  --stdin        Read names from stdin, one per line")
	fmt.Println("  --workers      Concurrent workers           [default: 30]")
	fmt.Println("  --stream       Stream NDJSON results as they arrive")
	fmt.Println("  --no-trademark Skip trademark search URLs in output")
	fmt.Println()
	fmt.Println("Check flags:")
	fmt.Println("  --tld              TLDs (comma-separated)       [default: com]")
	fmt.Println("  --workers          Concurrent workers           [default: 50]")
	fmt.Println("  --format           Export: txt,json,csv         [default: txt]")
	fmt.Println("  --regex, -r        Regex filter for domain prefix")
	fmt.Println("  --delay            Delay between queries (ms)   [default: 0]")
	fmt.Println("  --show-registered  Also save registered domains to file")
	fmt.Println()
	fmt.Println("TLDs flags:")
	fmt.Println("  --rdap     Only TLDs with RDAP support")
	fmt.Println("  --country  Only country-code TLDs")
	fmt.Println("  --refresh  Force refresh of TLD cache")
}

func runPresets() {
	fmt.Println(titleStyle.Render("◆ domh presets"))
	fmt.Println()
	// Sort keys for stable output
	keys := make([]string, 0, len(presets.Presets))
	for k := range presets.Presets {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, name := range keys {
		tlds := presets.Presets[name]
		fmt.Printf("  %-12s %s\n", name, dimStyle.Render(strings.Join(tlds, ", ")))
	}
	fmt.Println()
}

func runMCP() {
	s := mcpserver.NewMCPServer("domh", version, mcpserver.WithToolCapabilities(false))

	s.AddTool(mcp.NewTool("check_domain",
		mcp.WithDescription("Check if a single domain name is available for registration. Returns availability, method, and pricing."),
		mcp.WithString("domain", mcp.Required(), mcp.Description("Full domain name to check (e.g. coolname.com, myapp.com.br)")),
	), mcpCheckDomain)

	s.AddTool(mcp.NewTool("check_domains",
		mcp.WithDescription("Check multiple full domain names for availability. Returns availability, pricing, and buy URLs for each."),
		mcp.WithString("domains", mcp.Required(), mcp.Description("Comma-separated list of full domains (e.g. cool.com,cool.dev,cool.com.br)")),
	), mcpCheckDomains)

	s.AddTool(mcp.NewTool("scan_names",
		mcp.WithDescription("Check one or more base names across multiple TLDs. Takes a list of names and TLD list or preset, returns all combinations with availability, pricing, and trademark search URLs. Best for AI-driven naming workflows."),
		mcp.WithString("names", mcp.Required(), mcp.Description("Comma-separated base names without TLD (e.g. kora,nexus,velo,plex)")),
		mcp.WithString("tlds", mcp.Description("Comma-separated TLDs (e.g. com,io,app,com.br). Omit if using preset.")),
		mcp.WithString("preset", mcp.Description("TLD preset: saas, startup, brazil, brazil-full, brazil-saas, global-saas, tech, popular, classic, enterprise, creative, ecommerce, finance, web, trendy, country, br-pro")),
		mcp.WithBoolean("trademark", mcp.Description("Include trademark search URLs (INPI, USPTO, EUIPO). Default: true")),
	), mcpScanNames)

	s.AddTool(mcp.NewTool("check_with_preset",
		mcp.WithDescription("Check a single base name across a curated TLD preset."),
		mcp.WithString("name", mcp.Required(), mcp.Description("Base name to check (without TLD)")),
		mcp.WithString("preset", mcp.Required(), mcp.Description("Preset: saas, startup, brazil, brazil-full, global-saas, tech, popular, classic, enterprise, creative, ecommerce, finance, web, trendy, country, br-pro")),
	), mcpCheckWithPreset)

	s.AddTool(mcp.NewTool("generate_names",
		mcp.WithDescription("Generate pronounceable domain name candidates by length and phonetic pattern. Use to build a candidate list before checking availability."),
		mcp.WithNumber("length", mcp.Required(), mcp.Description("Name length: 3, 4, or 5 characters")),
		mcp.WithString("pattern", mcp.Description("Phonetic pattern: CVC, VCV, CVCV, CVCVC, ALL. Default: ALL")),
		mcp.WithString("tld", mcp.Description("TLD to append for preview. Default: com")),
	), mcpGenerateNames)

	s.AddTool(mcp.NewTool("get_prices",
		mcp.WithDescription("Get registrar prices and buy links for a domain. Covers 20+ registrars including Registro.br for .br TLDs."),
		mcp.WithString("domain", mcp.Required(), mcp.Description("Full domain name (e.g. coolname.com, myapp.com.br)")),
	), mcpGetPrices)

	s.AddTool(mcp.NewTool("trademark_urls",
		mcp.WithDescription("Get trademark search URLs for a name across INPI (Brazil), USPTO (US), and EUIPO (EU). Use to check if a candidate name is already trademarked."),
		mcp.WithString("name", mcp.Required(), mcp.Description("Base name to search trademarks for (e.g. kora, nexus)")),
	), mcpTrademarkURLs)

	s.AddTool(mcp.NewTool("list_presets",
		mcp.WithDescription("List all available TLD presets with their TLD lists."),
	), mcpListPresets)

	if err := mcpserver.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "MCP server error: %v\n", err)
		os.Exit(1)
	}
}

func mcpCheckDomain(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

func mcpCheckDomains(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

func mcpCheckWithPreset(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

func mcpGenerateNames(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
	out := domains
	if len(out) > 100 {
		out = out[:100]
	}
	type result struct {
		Count   int      `json:"count"`
		Domains []string `json:"domains"`
	}
	b, _ := json.Marshal(result{Count: len(domains), Domains: out})
	return mcp.NewToolResultText(string(b)), nil
}

func mcpGetPrices(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	domain, err := req.RequireString("domain")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	pr := pricing.GetPrices(strings.TrimSpace(strings.ToLower(domain)))
	b, _ := json.Marshal(pr)
	return mcp.NewToolResultText(string(b)), nil
}

func mcpListPresets(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	b, _ := json.Marshal(presets.List())
	return mcp.NewToolResultText(string(b)), nil
}

func mcpScanNames(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	rawNames, err := req.RequireString("names")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var names []string
	for _, n := range strings.Split(rawNames, ",") {
		n = strings.TrimSpace(strings.ToLower(n))
		if n != "" {
			names = append(names, n)
		}
	}
	if len(names) == 0 {
		return mcp.NewToolResultError("no valid names provided"), nil
	}

	// Resolve TLDs from preset or explicit list
	tlds := []string{"com", "io", "app"}
	if rawTLDs := req.GetString("tlds", ""); rawTLDs != "" {
		tlds = nil
		for _, t := range strings.Split(rawTLDs, ",") {
			if t = strings.TrimSpace(strings.ToLower(t)); t != "" {
				tlds = append(tlds, t)
			}
		}
	} else if presetName := req.GetString("preset", ""); presetName != "" {
		p, ok := presets.Get(presetName)
		if !ok {
			return mcp.NewToolResultError(fmt.Sprintf("unknown preset: %s", presetName)), nil
		}
		tlds = p
	}

	includeTrademark := req.GetBool("trademark", true)

	// Build domain list
	var domains []string
	for _, name := range names {
		for _, tld := range tlds {
			domains = append(domains, fmt.Sprintf("%s.%s", name, tld))
		}
	}

	results := scanner.CheckMultiple(domains)

	type domainResult struct {
		Domain    string                   `json:"domain"`
		Name      string                   `json:"name"`
		TLD       string                   `json:"tld"`
		Available bool                     `json:"available"`
		Error     bool                     `json:"error,omitempty"`
		Method    string                   `json:"method,omitempty"`
		Pricing   *pricing.PriceResult     `json:"pricing,omitempty"`
		Trademark *trademark.SearchURLs    `json:"trademark,omitempty"`
	}

	type scanResult struct {
		Query struct {
			Names []string `json:"names"`
			TLDs  []string `json:"tlds"`
		} `json:"query"`
		Available []domainResult `json:"available"`
		Taken     []domainResult `json:"taken"`
		Errors    []domainResult `json:"errors"`
		Stats     struct {
			Checked   int `json:"checked"`
			Available int `json:"available"`
			Taken     int `json:"taken"`
			Errors    int `json:"errors"`
		} `json:"stats"`
	}

	var out scanResult
	out.Query.Names = names
	out.Query.TLDs = tlds

	// Deduplicate trademark URLs per name
	tmCache := make(map[string]*trademark.SearchURLs)
	if includeTrademark {
		for _, name := range names {
			urls := trademark.For(name)
			tmCache[name] = &urls
		}
	}

	for _, r := range results {
		// Extract base name: everything before first dot
		baseName := strings.SplitN(r.Domain, ".", 2)[0]
		if r.Available {
			pr := pricing.GetPrices(r.Domain)
			dr := domainResult{
				Domain:    r.Domain,
				Name:      baseName,
				TLD:       r.TLD,
				Available: true,
				Method:    r.Method,
				Pricing:   &pr,
				Trademark: tmCache[baseName],
			}
			out.Available = append(out.Available, dr)
		} else if r.Error {
			out.Errors = append(out.Errors, domainResult{Domain: r.Domain, Name: baseName, TLD: r.TLD, Error: true})
		} else {
			out.Taken = append(out.Taken, domainResult{Domain: r.Domain, Name: baseName, TLD: r.TLD, Available: false})
		}
	}

	out.Stats.Checked = len(results)
	out.Stats.Available = len(out.Available)
	out.Stats.Taken = len(out.Taken)
	out.Stats.Errors = len(out.Errors)

	b, _ := json.Marshal(out)
	return mcp.NewToolResultText(string(b)), nil
}

func mcpTrademarkURLs(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, err := req.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	name = strings.TrimSpace(strings.ToLower(name))
	urls := trademark.For(name)
	type result struct {
		Name      string               `json:"name"`
		Trademark trademark.SearchURLs `json:"trademark"`
	}
	b, _ := json.Marshal(result{Name: name, Trademark: urls})
	return mcp.NewToolResultText(string(b)), nil
}

func runCheck(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: domh check <words-file> [--tld com,dev,io]")
		os.Exit(1)
	}

	file := args[0]
	tlds := []string{"com"}
	workers := 50
	formats := []export.Format{export.FormatTXT}
	regexFilter := ""
	delayMs := 0
	showRegistered := false

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--tld":
			if i+1 < len(args) {
				tlds = strings.Split(args[i+1], ",")
				i++
			}
		case "--workers":
			if i+1 < len(args) {
				w, _ := strconv.Atoi(args[i+1])
				if w > 0 {
					workers = w
				}
				i++
			}
		case "--format":
			if i+1 < len(args) {
				formats = nil
				for _, f := range strings.Split(args[i+1], ",") {
					formats = append(formats, export.Format(strings.TrimSpace(f)))
				}
				i++
			}
		case "--regex", "-r":
			if i+1 < len(args) {
				regexFilter = args[i+1]
				i++
			}
		case "--delay":
			if i+1 < len(args) {
				d, _ := strconv.Atoi(args[i+1])
				if d > 0 {
					delayMs = d
				}
				i++
			}
		case "--show-registered":
			showRegistered = true
		}
	}

	f, err := os.Open(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot open file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	var domains []string
	s := bufio.NewScanner(f)
	for s.Scan() {
		word := strings.TrimSpace(s.Text())
		if word == "" || strings.HasPrefix(word, "#") {
			continue
		}
		for _, tld := range tlds {
			domains = append(domains, fmt.Sprintf("%s.%s", strings.ToLower(word), tld))
		}
	}

	if len(domains) == 0 {
		fmt.Fprintln(os.Stderr, "No words found in file.")
		os.Exit(1)
	}

	domains = applyRegexFilter(domains, regexFilter)

	if !confirmLargeScan(len(domains), workers, false, false) {
		return
	}

	delay := time.Duration(delayMs) * time.Millisecond
	startScanWithDomains(domains, tlds, workers, formats, "dict", delay, showRegistered, false, false)
}

// suggestResult is the JSON schema for `domh suggest` output.
type suggestResult struct {
	Query struct {
		Names []string `json:"names"`
		TLDs  []string `json:"tlds"`
	} `json:"query"`
	Available []suggestDomain `json:"available"`
	Taken     []suggestDomain `json:"taken"`
	Errors    []suggestDomain `json:"errors"`
	Stats     struct {
		Checked     int   `json:"checked"`
		Available   int   `json:"available"`
		Taken       int   `json:"taken"`
		Errors      int   `json:"errors"`
		ElapsedMs   int64 `json:"elapsed_ms"`
	} `json:"stats"`
}

type suggestDomain struct {
	Domain    string               `json:"domain"`
	Name      string               `json:"name"`
	TLD       string               `json:"tld"`
	Available bool                 `json:"available"`
	Method    string               `json:"method,omitempty"`
	Pricing   *pricing.PriceResult `json:"pricing,omitempty"`
	Trademark *trademark.SearchURLs `json:"trademark,omitempty"`
}

func runSuggest(args []string) {
	tlds := []string{"com", "io", "app"}
	tldSet := false
	presetName := ""
	workers := 30
	fromStdin := false
	stream := false
	includeTrademark := true
	var names []string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--tld":
			if i+1 < len(args) {
				tlds = nil
				for _, t := range strings.Split(args[i+1], ",") {
					if t = strings.TrimSpace(t); t != "" {
						tlds = append(tlds, t)
					}
				}
				tldSet = true
				i++
			}
		case "--preset":
			if i+1 < len(args) {
				presetName = args[i+1]
				i++
			}
		case "--workers":
			if i+1 < len(args) {
				if w, err := strconv.Atoi(args[i+1]); err == nil && w > 0 {
					workers = w
				}
				i++
			}
		case "--stdin":
			fromStdin = true
		case "--stream":
			stream = true
		case "--no-trademark":
			includeTrademark = false
		default:
			if !strings.HasPrefix(args[i], "-") {
				names = append(names, strings.ToLower(strings.TrimSpace(args[i])))
			}
		}
	}

	// Auto-detect non-TTY stdin piping
	if !fromStdin {
		if fi, _ := os.Stdin.Stat(); fi != nil && (fi.Mode()&os.ModeCharDevice) == 0 {
			fromStdin = true
		}
	}

	// Read names from stdin
	if fromStdin {
		sc := bufio.NewScanner(os.Stdin)
		for sc.Scan() {
			if n := strings.ToLower(strings.TrimSpace(sc.Text())); n != "" {
				names = append(names, n)
			}
		}
	}

	if len(names) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: domh suggest <name1> [name2...] [--tld com,io,app] [--preset saas]")
		fmt.Fprintln(os.Stderr, "       echo -e 'kora\\nnexus' | domh suggest --preset global-saas")
		os.Exit(1)
	}

	// Resolve TLDs
	if presetName != "" {
		p, ok := presets.Get(presetName)
		if !ok {
			fmt.Fprintf(os.Stderr, "Unknown preset: %s\nRun 'domh presets' to see available presets.\n", presetName)
			os.Exit(1)
		}
		tlds = p
	} else if !tldSet {
		tlds = []string{"com", "io", "app"}
	}

	// Build domain list
	var domains []string
	for _, name := range names {
		for _, tld := range tlds {
			domains = append(domains, fmt.Sprintf("%s.%s", name, tld))
		}
	}

	// Trademark URL cache (one per unique name)
	tmCache := make(map[string]*trademark.SearchURLs)
	if includeTrademark {
		for _, name := range names {
			urls := trademark.For(name)
			tmCache[name] = &urls
		}
	}

	start := time.Now()

	if stream {
		// NDJSON streaming: emit each result as it arrives
		s := scanner.NewWithDelay(workers, 0)
		s.OnResult = func(r scanner.Result) {
			baseName := strings.SplitN(r.Domain, ".", 2)[0]
			sd := suggestDomain{
				Domain:    r.Domain,
				Name:      baseName,
				TLD:       r.TLD,
				Available: r.Available,
				Method:    r.Method,
				Trademark: tmCache[baseName],
			}
			if r.Available {
				pr := pricing.GetPrices(r.Domain)
				sd.Pricing = &pr
			}
			b, _ := json.Marshal(sd)
			fmt.Println(string(b))
		}
		s.Run(domains)
		<-s.Done
		return
	}

	// Collect all results, then output summary JSON
	results := scanner.CheckMultiple(domains)

	var out suggestResult
	out.Query.Names = names
	out.Query.TLDs = tlds
	out.Available = []suggestDomain{}
	out.Taken = []suggestDomain{}
	out.Errors = []suggestDomain{}

	for _, r := range results {
		baseName := strings.SplitN(r.Domain, ".", 2)[0]
		sd := suggestDomain{
			Domain:    r.Domain,
			Name:      baseName,
			TLD:       r.TLD,
			Available: r.Available,
			Method:    r.Method,
			Trademark: tmCache[baseName],
		}
		if r.Available {
			pr := pricing.GetPrices(r.Domain)
			sd.Pricing = &pr
		}
		switch {
		case r.Error:
			out.Errors = append(out.Errors, sd)
		case r.Available:
			out.Available = append(out.Available, sd)
		default:
			out.Taken = append(out.Taken, sd)
		}
	}

	out.Stats.Checked = len(results)
	out.Stats.Available = len(out.Available)
	out.Stats.Taken = len(out.Taken)
	out.Stats.Errors = len(out.Errors)
	out.Stats.ElapsedMs = time.Since(start).Milliseconds()

	b, _ := json.MarshalIndent(out, "", "  ")
	fmt.Println(string(b))
}

func runTLDs(args []string) {
	rdapOnly := false
	countryOnly := false
	refresh := false

	for _, arg := range args {
		switch arg {
		case "--rdap":
			rdapOnly = true
		case "--country":
			countryOnly = true
		case "--refresh":
			refresh = true
		}
	}

	if refresh {
		fmt.Print("Refreshing TLD cache... ")
		if err := registry.RefreshCache(); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("done.")
	}

	tlds, err := registry.GetCachedTLDs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching TLDs: %v\n", err)
		os.Exit(1)
	}

	var filtered []registry.TLDInfo
	for _, t := range tlds {
		if rdapOnly && t.RDAPUrl == "" {
			continue
		}
		if countryOnly && t.Type != "country-code" {
			continue
		}
		filtered = append(filtered, t)
	}

	for _, t := range filtered {
		suffix := ""
		if t.RDAPUrl != "" {
			suffix = dimStyle.Render(" (rdap)")
		}
		fmt.Printf("  .%-10s %s%s\n", t.Name, dimStyle.Render(t.Type), suffix)
	}
	fmt.Printf("\n  Total: %d TLDs\n", len(filtered))
}

func runUpdate() {
	var latest *selfupdate.Release
	var found bool
	var detectErr error

	err := spinner.New().
		Title("Checking for updates...").
		Action(func() {
			latest, found, detectErr = selfupdate.DetectLatest(context.Background(), selfupdate.ParseSlug("codeatlasdev/domain-hunter"))
		}).
		Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Spinner error: %v\n", err)
		os.Exit(1)
	}
	if detectErr != nil {
		fmt.Fprintf(os.Stderr, "Error checking for updates: %v\n", detectErr)
		os.Exit(1)
	}
	if !found {
		fmt.Println("No releases found.")
		return
	}

	if latest.LessOrEqual(version) {
		fmt.Printf("Already up to date (%s).\n", version)
		return
	}

	fmt.Printf("New version available: %s → %s\n", version, latest.Version())

	exe, err := selfupdate.ExecutablePath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not locate executable: %v\n", err)
		os.Exit(1)
	}

	var updateErr error
	err = spinner.New().
		Title(fmt.Sprintf("Updating to %s...", latest.Version())).
		Action(func() {
			updateErr = selfupdate.DefaultUpdater().UpdateTo(context.Background(), latest, exe)
		}).
		Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Spinner error: %v\n", err)
		os.Exit(1)
	}
	if updateErr != nil {
		fmt.Fprintf(os.Stderr, "Update failed: %v\n", updateErr)
		os.Exit(1)
	}

	fmt.Printf("✓ Updated to %s\n", latest.Version())
}

func runInteractive() {
	cfg, err := wizard.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cancelled.\n")
		os.Exit(0)
	}

	domains := scanner.GenerateDomains(cfg.Length, cfg.Pattern, cfg.TLDs)
	if len(domains) == 0 {
		fmt.Fprintln(os.Stderr, "No domains generated. Check length/pattern combination.")
		os.Exit(1)
	}

	if !confirmLargeScan(len(domains), cfg.Workers, false, false) {
		return
	}

	startScanWithDomains(domains, cfg.TLDs, cfg.Workers, cfg.Formats, string(cfg.Pattern), 0, false, false, false)
}

func runCLI(args []string) {
	tlds := []string{"com"}
	tldSet := false
	length := 3
	pattern := scanner.PatternAll
	workers := 50
	formats := []export.Format{export.FormatTXT}
	regexFilter := ""
	delayMs := 0
	showRegistered := false
	presetName := ""
	allTLDs := false
	prefixes := []string{}
	suffixes := []string{}
	info := false
	dryRun := false
	yes := false
	force := false
	batch := false
	jsonOutput := false
	baseName := ""

	// Parse args — collect positional (base name) and flags
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--tld":
			if i+1 < len(args) {
				tlds = strings.Split(args[i+1], ",")
				tldSet = true
				i++
			}
		case "--preset":
			if i+1 < len(args) {
				presetName = args[i+1]
				i++
			}
		case "--all":
			allTLDs = true
		case "--length":
			if i+1 < len(args) {
				l, _ := strconv.Atoi(args[i+1])
				if l >= 3 && l <= 5 {
					length = l
				}
				i++
			}
		case "--pattern":
			if i+1 < len(args) {
				pattern = scanner.Pattern(strings.ToUpper(args[i+1]))
				i++
			}
		case "--prefix":
			if i+1 < len(args) {
				prefixes = strings.Split(args[i+1], ",")
				i++
			}
		case "--suffix":
			if i+1 < len(args) {
				suffixes = strings.Split(args[i+1], ",")
				i++
			}
		case "--workers":
			if i+1 < len(args) {
				w, _ := strconv.Atoi(args[i+1])
				if w > 0 {
					workers = w
				}
				i++
			}
		case "--format":
			if i+1 < len(args) {
				formats = nil
				for _, f := range strings.Split(args[i+1], ",") {
					formats = append(formats, export.Format(strings.TrimSpace(f)))
				}
				i++
			}
		case "--regex", "-r":
			if i+1 < len(args) {
				regexFilter = args[i+1]
				i++
			}
		case "--delay":
			if i+1 < len(args) {
				d, _ := strconv.Atoi(args[i+1])
				if d > 0 {
					delayMs = d
				}
				i++
			}
		case "--info":
			info = true
		case "--show-registered":
			showRegistered = true
		case "--dry-run":
			dryRun = true
		case "--yes", "-y":
			yes = true
		case "--force":
			force = true
		case "--batch", "--no-tui":
			batch = true
		case "--json":
			jsonOutput = true
			batch = true
		default:
			if !strings.HasPrefix(args[i], "-") && baseName == "" {
				baseName = args[i]
			}
		}
	}

	// Auto-detect non-TTY → batch mode
	if !batch {
		if fi, _ := os.Stdout.Stat(); fi != nil && (fi.Mode()&os.ModeCharDevice) == 0 {
			batch = true
		}
	}

	// Resolve TLDs: --all > --preset > --tld
	if allTLDs {
		cached, err := registry.GetCachedTLDs()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching TLDs: %v\n", err)
			os.Exit(1)
		}
		tlds = make([]string, 0, len(cached))
		for _, t := range cached {
			tlds = append(tlds, t.Name)
		}
	} else if presetName != "" {
		p, ok := presets.Get(presetName)
		if !ok {
			fmt.Fprintf(os.Stderr, "Unknown preset: %s\nRun 'domh presets' to see available presets.\n", presetName)
			os.Exit(1)
		}
		tlds = p
	} else if !tldSet {
		tlds = []string{"com"}
	}

	// Generate domains
	var domains []string

	if baseName != "" && (len(prefixes) > 0 || len(suffixes) > 0) {
		// Prefix/suffix combo mode
		var names []string
		for _, p := range prefixes {
			names = append(names, p+baseName)
		}
		for _, s := range suffixes {
			names = append(names, baseName+s)
		}
		if len(names) == 0 {
			names = []string{baseName}
		}
		for _, tld := range tlds {
			for _, name := range names {
				domains = append(domains, fmt.Sprintf("%s.%s", name, tld))
			}
		}
	} else if baseName != "" {
		// Simple base name mode (no pattern generation)
		for _, tld := range tlds {
			domains = append(domains, fmt.Sprintf("%s.%s", baseName, tld))
		}
	} else {
		// Pattern generation mode
		domains = scanner.GenerateDomains(length, pattern, tlds)
	}

	if len(domains) == 0 {
		fmt.Fprintln(os.Stderr, "No domains generated. Check length/pattern combination.")
		os.Exit(1)
	}

	domains = applyRegexFilter(domains, regexFilter)

	// Dry run
	if dryRun {
		fmt.Printf("Dry run — %d domains would be checked:\n", len(domains))
		for _, d := range domains {
			fmt.Printf("  %s\n", d)
		}
		return
	}

	// Performance warning
	if !force && !confirmLargeScan(len(domains), workers, yes, force) {
		return
	}

	delay := time.Duration(delayMs) * time.Millisecond

	if batch {
		startBatchScan(domains, tlds, workers, formats, delay, showRegistered, info, jsonOutput)
	} else {
		startScanWithDomains(domains, tlds, workers, formats, string(pattern), delay, showRegistered, info, false)
	}
}

func startBatchScan(domains []string, tlds []string, workers int, formats []export.Format, delay time.Duration, showRegistered bool, info bool, jsonOutput bool) {
	exp, err := export.NewWithOptions(formats, showRegistered)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Export error: %v\n", err)
		os.Exit(1)
	}
	defer exp.Close()

	sc := scanner.NewWithDelay(workers, delay)

	sc.OnResult = func(r scanner.Result) {
		exp.Append(r)

		if jsonOutput {
			type jsonLine struct {
				Domain    string               `json:"domain"`
				Available bool                 `json:"available"`
				Error     bool                 `json:"error,omitempty"`
				Method    string               `json:"method,omitempty"`
				TLD       string               `json:"tld,omitempty"`
				Pricing   *pricing.PriceResult `json:"pricing,omitempty"`
			}
			jl := jsonLine{Domain: r.Domain, Available: r.Available, Error: r.Error, Method: r.Method, TLD: r.TLD}
			if r.Available {
				pr := pricing.GetPrices(r.Domain)
				jl.Pricing = &pr
			}
			b, _ := json.Marshal(jl)
			fmt.Println(string(b))
			return
		}

		status := "TAKEN"
		if r.Available {
			status = "AVAILABLE"
		} else if r.Error {
			status = "ERROR"
		}
		fmt.Printf("%s %s\n", status, r.Domain)

		if info && !r.Available && !r.Error {
			parts := strings.SplitN(r.Domain, ".", 2)
			if len(parts) == 2 {
				if di := scanner.FetchDomainInfo(r.Domain, parts[1]); di != nil {
					fmt.Printf("  Registrar: %s  Created: %s  Expires: %s\n", di.Registrar, di.CreatedDate, di.ExpiryDate)
				}
			}
		}
	}

	sc.Run(domains)
	<-sc.Done
}

func applyRegexFilter(domains []string, regexFilter string) []string {
	if regexFilter == "" {
		return domains
	}
	re := regexp.MustCompile(regexFilter)
	var filtered []string
	for _, d := range domains {
		name := strings.SplitN(d, ".", 2)[0]
		if re.MatchString(name) {
			filtered = append(filtered, d)
		}
	}
	return filtered
}

func confirmLargeScan(count, workers int, yes, force bool) bool {
	if force || count <= 10000 {
		return true
	}

	if yes {
		return true
	}

	// Non-interactive: skip prompt
	fi, _ := os.Stdin.Stat()
	if fi != nil && (fi.Mode()&os.ModeCharDevice) == 0 {
		return true
	}

	rate := float64(workers) * 6
	if rate > 300 {
		rate = 300
	}
	etaSec := float64(count) / rate
	eta := time.Duration(etaSec) * time.Second

	fmt.Println()
	fmt.Println(warnStyle.Render("⚠ Large scan detected"))
	fmt.Printf("  Domains: %s\n", formatNumber(count))
	fmt.Printf("  Estimated time: ~%s (at %.0f/s)\n", eta.Round(time.Second), rate)
	fmt.Printf("  Network requests: ~%s\n", formatNumber(count))
	fmt.Println()
	fmt.Print("  Continue? [Y/n] ")

	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))
	if answer == "n" || answer == "no" {
		fmt.Println("  Cancelled.")
		return false
	}
	return true
}

func formatNumber(n int) string {
	s := strconv.Itoa(n)
	if len(s) <= 3 {
		return s
	}
	var result []byte
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, byte(c))
	}
	return string(result)
}

func startScanWithDomains(domains []string, tlds []string, workers int, formats []export.Format, pattern string, delay time.Duration, showRegistered bool, info bool, _ bool) {
	exp, err := export.NewWithOptions(formats, showRegistered)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Export error: %v\n", err)
		os.Exit(1)
	}
	defer exp.Close()

	sc := scanner.NewWithDelay(workers, delay)

	originalOnResult := sc.OnResult
	sc.OnResult = func(r scanner.Result) {
		exp.Append(r)
		if originalOnResult != nil {
			originalOnResult(r)
		}
	}

	cfg := tui.Config{
		TLDs:    tlds,
		Length:  0,
		Pattern: pattern,
		Workers: workers,
	}
	model := tui.NewModel(sc, cfg)

	sc.Run(domains)

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
		os.Exit(1)
	}

	// Final summary
	stats := sc.Stats()
	results := sc.Results()
	var available []string
	for _, r := range results {
		if r.Available {
			available = append(available, r.Domain)
		}
	}
	sort.Strings(available)

	fmt.Println()
	fmt.Println(titleStyle.Render("◆ Domain Hunter — Complete"))
	fmt.Println()
	elapsed := time.Since(stats.StartTime).Round(time.Second)
	fmt.Printf("  Checked: %d │ Available: %s │ Errors: %d │ Time: %s\n",
		stats.Checked,
		greenBold.Render(fmt.Sprintf("%d", len(available))),
		stats.Errors,
		elapsed,
	)
	fmt.Println()

	if len(available) > 0 {
		fmt.Println(greenBold.Render("  Available domains:"))
		fmt.Println()
		for _, d := range available {
			fmt.Printf("    %s %s\n", greenBold.Render("✓"), greenBold.Render(d))
			for _, r := range results {
				if r.Domain == d && r.Pricing != nil && len(r.Pricing.Prices) > 0 {
					for i, p := range r.Pricing.Prices {
						if i >= 5 {
							break
						}
						marker := "  "
						if i == 0 {
							marker = "→ "
						}
						fmt.Printf("      %s%-12s $%.2f", marker, p.Registrar, p.RegisterPrice)
						if p.BuyURL != "" {
							fmt.Printf("  %s", dimStyle.Render(p.BuyURL))
						}
						fmt.Println()
					}
					fmt.Println()
					break
				}
			}
		}
	}

	// Show info for taken domains if --info
	if info {
		var taken []scanner.Result
		for _, r := range results {
			if !r.Available && !r.Error {
				taken = append(taken, r)
			}
		}
		if len(taken) > 0 {
			fmt.Println(redBold.Render("  Taken domains:"))
			fmt.Println()
			for _, r := range taken {
				fmt.Printf("    %s %s\n", redBold.Render("✗"), r.Domain)
				parts := strings.SplitN(r.Domain, ".", 2)
				if len(parts) == 2 {
					if di := scanner.FetchDomainInfo(r.Domain, parts[1]); di != nil {
						fmt.Printf("      Registrar: %s  Created: %s  Expires: %s\n", di.Registrar, di.CreatedDate, di.ExpiryDate)
					}
				}
			}
			fmt.Println()
		}
	}

	fmt.Println(dimStyle.Render(fmt.Sprintf("  Saved to: %s", strings.Join(exp.Filenames(), ", "))))
	fmt.Println()
}
