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
	fmt.Println("  domh                       Interactive wizard")
	fmt.Println("  domh scan [name] [flags]   Scan domains")
	fmt.Println("  domh check <file> [flags]  Dictionary mode")
	fmt.Println("  domh tlds [flags]          List TLDs")
	fmt.Println("  domh presets               List presets")
	fmt.Println("  domh mcp                   Start MCP server (stdio)")
	fmt.Println("  domh update                Self-update")
	fmt.Println("  domh version               Version info")
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
	fmt.Println("  --batch        Plain output (no TUI, CI-friendly)")
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
		mcp.WithDescription("Check if a domain name is available for registration"),
		mcp.WithString("domain", mcp.Required(), mcp.Description("Full domain name to check (e.g. coolname.com)")),
	), mcpCheckDomain)

	s.AddTool(mcp.NewTool("check_domains",
		mcp.WithDescription("Check multiple domain names for availability"),
		mcp.WithString("domains", mcp.Required(), mcp.Description("Comma-separated list of domains to check")),
	), mcpCheckDomains)

	s.AddTool(mcp.NewTool("check_with_preset",
		mcp.WithDescription("Check a name across a curated set of TLDs"),
		mcp.WithString("name", mcp.Required(), mcp.Description("Base name to check (without TLD)")),
		mcp.WithString("preset", mcp.Required(), mcp.Description("Preset name: startup, tech, creative, ecommerce, finance, popular, classic, enterprise, web, trendy, country, brazil")),
	), mcpCheckWithPreset)

	s.AddTool(mcp.NewTool("generate_names",
		mcp.WithDescription("Generate pronounceable domain names by pattern"),
		mcp.WithNumber("length", mcp.Required(), mcp.Description("Name length: 3, 4, or 5")),
		mcp.WithString("pattern", mcp.Description("Pattern: CVC, VCV, CVCV, CVCVC, ALL. Default: ALL")),
		mcp.WithString("tld", mcp.Description("TLD to append. Default: com")),
	), mcpGenerateNames)

	s.AddTool(mcp.NewTool("get_prices",
		mcp.WithDescription("Get registrar prices and buy links for a domain"),
		mcp.WithString("domain", mcp.Required(), mcp.Description("Domain to get prices for (e.g. coolname.com)")),
	), mcpGetPrices)

	s.AddTool(mcp.NewTool("list_presets",
		mcp.WithDescription("List all available TLD presets with their TLDs"),
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
		case "--batch":
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
		startBatchScan(domains, tlds, workers, formats, delay, showRegistered, info)
	} else {
		startScanWithDomains(domains, tlds, workers, formats, string(pattern), delay, showRegistered, info, false)
	}
}

func startBatchScan(domains []string, tlds []string, workers int, formats []export.Format, delay time.Duration, showRegistered bool, info bool) {
	exp, err := export.NewWithOptions(formats, showRegistered)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Export error: %v\n", err)
		os.Exit(1)
	}
	defer exp.Close()

	sc := scanner.NewWithDelay(workers, delay)

	sc.OnResult = func(r scanner.Result) {
		exp.Append(r)

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
