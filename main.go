package main

import (
	"bufio"
	"context"
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

	"github.com/codeatlasdev/domain-hunter/internal/export"
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
	fmt.Println("  domh                  Interactive wizard")
	fmt.Println("  domh scan [flags]     Generate & scan pronounceable domains")
	fmt.Println("  domh check <file>     Dictionary mode — scan words from file")
	fmt.Println("  domh tlds [flags]     List available TLDs")
	fmt.Println("  domh update           Self-update to latest version")
	fmt.Println("  domh version          Show version info")
	fmt.Println()
	fmt.Println("Scan flags:")
	fmt.Println("  --tld              TLDs (comma-separated)       [default: com]")
	fmt.Println("  --length           Domain length (3-5)          [default: 3]")
	fmt.Println("  --pattern          CVC, VCV, CVCV, ALL          [default: ALL]")
	fmt.Println("  --workers          Concurrent workers           [default: 50]")
	fmt.Println("  --format           Export: txt,json,csv         [default: txt]")
	fmt.Println("  --regex, -r        Regex filter for domain prefix")
	fmt.Println("  --delay            Delay between queries (ms)   [default: 0]")
	fmt.Println("  --show-registered  Also save registered domains to file")
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

	// Read words from file
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

	if !confirmLargeScan(len(domains), workers) {
		return
	}

	delay := time.Duration(delayMs) * time.Millisecond
	startScanWithDomains(domains, tlds, workers, formats, "dict", delay, showRegistered)
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

	if !confirmLargeScan(len(domains), cfg.Workers) {
		return
	}

	startScanWithDomains(domains, cfg.TLDs, cfg.Workers, cfg.Formats, string(cfg.Pattern), 0, false)
}

func runCLI(args []string) {
	tlds := []string{"com"}
	length := 3
	pattern := scanner.PatternAll
	workers := 50
	formats := []export.Format{export.FormatTXT}
	regexFilter := ""
	delayMs := 0
	showRegistered := false

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--tld":
			if i+1 < len(args) {
				tlds = strings.Split(args[i+1], ",")
				i++
			}
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

	domains := scanner.GenerateDomains(length, pattern, tlds)
	if len(domains) == 0 {
		fmt.Fprintln(os.Stderr, "No domains generated. Check length/pattern combination.")
		os.Exit(1)
	}

	domains = applyRegexFilter(domains, regexFilter)

	delay := time.Duration(delayMs) * time.Millisecond
	startScanWithDomains(domains, tlds, workers, formats, string(pattern), delay, showRegistered)
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

// confirmLargeScan shows a warning for scans > 10000 domains. Returns false if user declines.
func confirmLargeScan(count, workers int) bool {
	if count <= 10000 {
		return true
	}

	// Only prompt in interactive mode (stdin is a terminal)
	fi, _ := os.Stdin.Stat()
	if fi != nil && (fi.Mode()&os.ModeCharDevice) == 0 {
		return true // piped input, skip prompt
	}

	rate := float64(workers) * 6 // ~6 domains/s per worker (conservative)
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

func startScanWithDomains(domains []string, tlds []string, workers int, formats []export.Format, pattern string, delay time.Duration, showRegistered bool) {
	// Init exporter
	exp, err := export.NewWithOptions(formats, showRegistered)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Export error: %v\n", err)
		os.Exit(1)
	}
	defer exp.Close()

	// Init scanner
	sc := scanner.NewWithDelay(workers, delay)

	// Wire export to scanner results
	originalOnResult := sc.OnResult
	sc.OnResult = func(r scanner.Result) {
		exp.Append(r)
		if originalOnResult != nil {
			originalOnResult(r)
		}
	}

	// Init TUI
	cfg := tui.Config{
		TLDs:    tlds,
		Length:  0,
		Pattern: pattern,
		Workers: workers,
	}
	model := tui.NewModel(sc, cfg)

	// Start scan
	sc.Run(domains)

	// Run TUI (blocks until done or quit)
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
			// Show all prices for this domain
			for _, r := range results {
				if r.Domain == d && r.Pricing != nil && len(r.Pricing.Prices) > 0 {
					for i, p := range r.Pricing.Prices {
						if i >= 5 { // top 5
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

	fmt.Println(dimStyle.Render(fmt.Sprintf("  Saved to: %s", strings.Join(exp.Filenames(), ", "))))
	fmt.Println()
}
