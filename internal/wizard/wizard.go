package wizard

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"

	"github.com/codeatlasdev/domain-hunter/internal/export"
	"github.com/codeatlasdev/domain-hunter/internal/registry"
	"github.com/codeatlasdev/domain-hunter/internal/scanner"
)

type Config struct {
	TLDs    []string
	Length  int
	Pattern scanner.Pattern
	Workers int
	Formats []export.Format
}

func Run() (*Config, error) {
	var (
		tldCategory string
		tlds        []string
		customTLDs  string
		length      string
		pattern     string
		workersStr  string
		formats     []string
	)

	// Count available TLDs
	totalTLDs := 0
	cached, _ := registry.GetCachedTLDs()
	if len(cached) > 0 {
		totalTLDs = len(cached)
	}

	totalLabel := ""
	if totalTLDs > 0 {
		totalLabel = fmt.Sprintf(" (%d available)", totalTLDs)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("TLD selection").
				Description("Which extensions to scan?").
				Options(
					huh.NewOption("Popular — com, net, org, dev, io, app, co, xyz", "popular"),
					huh.NewOption("Country — br, us, uk, de, fr, jp, au, ca", "country"),
					huh.NewOption("All"+totalLabel, "all"),
					huh.NewOption("Custom — type manually", "custom"),
				).
				Value(&tldCategory),
		),
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select TLDs").
				Description("Pick the extensions you want").
				Options(
					huh.NewOption(".com", "com").Selected(true),
					huh.NewOption(".net", "net"),
					huh.NewOption(".org", "org"),
					huh.NewOption(".dev", "dev"),
					huh.NewOption(".io", "io"),
					huh.NewOption(".app", "app"),
					huh.NewOption(".co", "co"),
					huh.NewOption(".xyz", "xyz"),
				).
				Value(&tlds).
				Validate(func(v []string) error {
					if len(v) == 0 {
						return fmt.Errorf("select at least one TLD")
					}
					return nil
				}),
		).WithHideFunc(func() bool { return tldCategory != "popular" }),
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select country TLDs").
				Description("Pick country-code extensions").
				Options(
					huh.NewOption(".br (Brazil)", "br").Selected(true),
					huh.NewOption(".us (United States)", "us"),
					huh.NewOption(".uk (United Kingdom)", "uk"),
					huh.NewOption(".de (Germany)", "de"),
					huh.NewOption(".fr (France)", "fr"),
					huh.NewOption(".jp (Japan)", "jp"),
					huh.NewOption(".au (Australia)", "au"),
					huh.NewOption(".ca (Canada)", "ca"),
					huh.NewOption(".in (India)", "in"),
					huh.NewOption(".it (Italy)", "it"),
				).
				Value(&tlds).
				Validate(func(v []string) error {
					if len(v) == 0 {
						return fmt.Errorf("select at least one TLD")
					}
					return nil
				}),
		).WithHideFunc(func() bool { return tldCategory != "country" }),
		huh.NewGroup(
			huh.NewInput().
				Title("Custom TLDs").
				Description("Comma-separated list (e.g. com,dev,br,xyz)").
				Placeholder("com,dev,io").
				Value(&customTLDs).
				Validate(func(v string) error {
					if strings.TrimSpace(v) == "" {
						return fmt.Errorf("enter at least one TLD")
					}
					return nil
				}),
		).WithHideFunc(func() bool { return tldCategory != "custom" }),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Domain length").
				Description("Number of characters in the domain name").
				Options(
					huh.NewOption("3 letters (bun, dev, kit)", "3"),
					huh.NewOption("4 letters (buno, kaze, tevo)", "4"),
					huh.NewOption("5 letters (nexus, pixel)", "5"),
				).
				Value(&length),
		),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Pattern").
				Description("Syllable structure for pronounceability").
				Options(
					huh.NewOption("CVC — consonant-vowel-consonant", "CVC"),
					huh.NewOption("VCV — vowel-consonant-vowel", "VCV"),
					huh.NewOption("CVCV — consonant-vowel-consonant-vowel", "CVCV"),
					huh.NewOption("CVCVC — consonant-vowel-consonant-vowel-consonant", "CVCVC"),
					huh.NewOption("ALL — all patterns for the length", "ALL"),
				).
				Value(&pattern),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Workers").
				Description("Number of concurrent goroutines").
				Placeholder("50").
				Value(&workersStr),
		),
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Export formats").
				Description("Results are saved in real-time").
				Options(
					huh.NewOption("txt (one domain per line)", "txt").Selected(true),
					huh.NewOption("json (array of objects)", "json"),
					huh.NewOption("csv (domain, tld, checked_at)", "csv"),
				).
				Value(&formats).
				Validate(func(v []string) error {
					if len(v) == 0 {
						return fmt.Errorf("select at least one format")
					}
					return nil
				}),
		),
	)

	if err := form.Run(); err != nil {
		return nil, err
	}

	// Resolve TLDs based on category
	switch tldCategory {
	case "all":
		tlds = nil
		for _, t := range cached {
			tlds = append(tlds, t.Name)
		}
		if len(tlds) == 0 {
			tlds = []string{"com"}
		}
	case "custom":
		tlds = nil
		for _, t := range strings.Split(customTLDs, ",") {
			t = strings.TrimSpace(strings.TrimPrefix(t, "."))
			if t != "" {
				tlds = append(tlds, strings.ToLower(t))
			}
		}
	}

	workers := 50
	if workersStr != "" {
		if w, err := strconv.Atoi(workersStr); err == nil && w > 0 {
			workers = w
		}
	}

	l, _ := strconv.Atoi(length)

	var fmts []export.Format
	for _, f := range formats {
		fmts = append(fmts, export.Format(f))
	}

	return &Config{
		TLDs:    tlds,
		Length:  l,
		Pattern: scanner.Pattern(pattern),
		Workers: workers,
		Formats: fmts,
	}, nil
}
