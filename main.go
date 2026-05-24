package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/codeatlasdev/domain-hunter/internal/export"
	"github.com/codeatlasdev/domain-hunter/internal/scanner"
	"github.com/codeatlasdev/domain-hunter/internal/tui"
	"github.com/codeatlasdev/domain-hunter/internal/wizard"
)

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#2563EB"))
	greenBold  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#10B981"))
	dimStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#64748B"))
)

func main() {
	if len(os.Args) < 2 {
		runInteractive()
		return
	}

	switch os.Args[1] {
	case "scan":
		runCLI(os.Args[2:])
	case "help", "--help", "-h":
		printHelp()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", os.Args[1])
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println(titleStyle.Render("◆ Domain Hunter") + " — bulk domain availability via RDAP")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  domain-hunter              Interactive wizard")
	fmt.Println("  domain-hunter scan [flags] Direct scan")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --tld      TLDs (comma-separated: com,dev,io)  [default: com]")
	fmt.Println("  --length   Domain length (3, 4, 5)             [default: 3]")
	fmt.Println("  --pattern  Pattern (CVC, VCV, CVCV, CVCVC, ALL) [default: ALL]")
	fmt.Println("  --workers  Concurrent workers                  [default: 50]")
	fmt.Println("  --format   Export formats (comma-separated)    [default: txt]")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  domain-hunter scan --tld com,dev --length 4 --workers 80 --format json")
	fmt.Println("  domain-hunter scan --tld com --length 3 --pattern cvc")
}

func runInteractive() {
	cfg, err := wizard.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cancelled.\n")
		os.Exit(0)
	}
	startScan(cfg.TLDs, cfg.Length, cfg.Pattern, cfg.Workers, cfg.Formats)
}

func runCLI(args []string) {
	tlds := []string{"com"}
	length := 3
	pattern := scanner.PatternAll
	workers := 50
	formats := []export.Format{export.FormatTXT}

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
		}
	}

	// Validate TLDs
	for _, tld := range tlds {
		if _, ok := scanner.Providers[tld]; !ok {
			fmt.Fprintf(os.Stderr, "Unsupported TLD: .%s\nSupported: %s\n", tld, strings.Join(scanner.SupportedTLDs, ", "))
			os.Exit(1)
		}
	}

	startScan(tlds, length, pattern, workers, formats)
}

func startScan(tlds []string, length int, pattern scanner.Pattern, workers int, formats []export.Format) {
	domains := scanner.GenerateDomains(length, pattern, tlds)

	if len(domains) == 0 {
		fmt.Fprintln(os.Stderr, "No domains generated. Check length/pattern combination.")
		os.Exit(1)
	}

	// Init exporter
	exp, err := export.New(formats)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Export error: %v\n", err)
		os.Exit(1)
	}
	defer exp.Close()

	// Init scanner
	sc := scanner.New(workers)

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
		Length:  length,
		Pattern: string(pattern),
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
		for _, d := range available {
			fmt.Printf("    ✓ %s\n", d)
		}
		fmt.Println()
	}

	fmt.Println(dimStyle.Render(fmt.Sprintf("  Saved to: %s", strings.Join(exp.Filenames(), ", "))))
	fmt.Println()
}
