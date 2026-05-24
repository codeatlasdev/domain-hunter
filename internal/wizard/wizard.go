package wizard

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/huh"

	"github.com/codeatlasdev/domain-hunter/internal/export"
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
		tlds       []string
		length     string
		pattern    string
		workersStr string
		formats    []string
	)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select TLDs").
				Description("Which extensions to scan?").
				Options(
					huh.NewOption(".com", "com").Selected(true),
					huh.NewOption(".net", "net"),
					huh.NewOption(".dev", "dev"),
					huh.NewOption(".io", "io"),
					huh.NewOption(".app", "app"),
					huh.NewOption(".org", "org"),
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
		),
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
