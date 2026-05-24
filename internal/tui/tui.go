package tui

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/codeatlasdev/domain-hunter/internal/scanner"
)

// Colors
var (
	blue     = lipgloss.Color("#2563EB")
	green    = lipgloss.Color("#10B981")
	red      = lipgloss.Color("#EF4444")
	amber    = lipgloss.Color("#F59E0B")
	cyan     = lipgloss.Color("#06B6D4")
	slate400 = lipgloss.Color("#94A3B8")
	slate500 = lipgloss.Color("#64748B")
	white    = lipgloss.Color("#F8FAFC")
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(blue)
	dimStyle   = lipgloss.NewStyle().Foreground(slate500)
	boldWhite  = lipgloss.NewStyle().Bold(true).Foreground(white)
	greenBold  = lipgloss.NewStyle().Bold(true).Foreground(green)
	redStyle   = lipgloss.NewStyle().Foreground(red)
	amberStyle = lipgloss.NewStyle().Foreground(amber)
	cyanStyle  = lipgloss.NewStyle().Foreground(cyan)

	headerStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(blue).
			Padding(0, 2).
			MarginBottom(1)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#334155")).
			Padding(0, 2)

	availBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(green).
			Padding(0, 2)
)

type Config struct {
	TLDs    []string
	Length  int
	Pattern string
	Workers int
}

type Model struct {
	spinner   spinner.Model
	scanner   *scanner.Scanner
	config    Config
	available []string
	logs      []string
	mu        sync.Mutex
	done      bool
}

type tickMsg struct{}
type doneMsg struct{}

func NewModel(sc *scanner.Scanner, cfg Config) *Model {
	s := spinner.New()
	s.Spinner = spinner.MiniDot
	s.Style = lipgloss.NewStyle().Foreground(blue)

	m := &Model{
		spinner: s,
		scanner: sc,
		config:  cfg,
	}

	sc.OnResult = func(r scanner.Result) {
		m.mu.Lock()
		defer m.mu.Unlock()
		ts := time.Now().Format("15:04:05")
		if r.Available {
			m.available = append(m.available, r.Domain)
			m.logs = append(m.logs, fmt.Sprintf("[%s] ✓ %s", ts, r.Domain))
		} else if r.Error {
			m.logs = append(m.logs, fmt.Sprintf("[%s] ⚠ %s (error)", ts, r.Domain))
		} else {
			m.logs = append(m.logs, fmt.Sprintf("[%s]   %s", ts, r.Domain))
		}
		if len(m.logs) > 300 {
			m.logs = m.logs[len(m.logs)-300:]
		}
	}

	return m
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, doTick(), waitDone(m.scanner.Done))
}

func doTick() tea.Cmd {
	return tea.Tick(150*time.Millisecond, func(time.Time) tea.Msg { return tickMsg{} })
}

func waitDone(ch chan struct{}) tea.Cmd {
	return func() tea.Msg {
		<-ch
		return doneMsg{}
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case tickMsg:
		return m, doTick()
	case doneMsg:
		m.done = true
		return m, tea.Quit
	}
	return m, nil
}

func (m *Model) View() string {
	stats := m.scanner.Stats()
	var b strings.Builder

	// Header
	tldStr := strings.Join(m.config.TLDs, ", .")
	header := headerStyle.Render(
		titleStyle.Render("◆ Domain Hunter") + "  " +
			dimStyle.Render(fmt.Sprintf("%d-letter .%s • %s • %d workers",
				m.config.Length, tldStr, m.config.Pattern, m.config.Workers)),
	)
	b.WriteString(header + "\n\n")

	// Progress
	checked := int(stats.Checked)
	total := stats.Total
	elapsed := time.Since(stats.StartTime)
	pct := 0
	if total > 0 {
		pct = (checked * 100) / total
	}
	rate := float64(0)
	if elapsed.Seconds() > 0 {
		rate = float64(checked) / elapsed.Seconds()
	}
	eta := "─"
	if rate > 0 && checked < total {
		remaining := float64(total-checked) / rate
		eta = (time.Duration(remaining) * time.Second).Round(time.Second).String()
	}

	// Bar
	barW := 40
	filled := (pct * barW) / 100
	bar := lipgloss.NewStyle().Foreground(blue).Render(strings.Repeat("━", filled)) +
		lipgloss.NewStyle().Foreground(lipgloss.Color("#334155")).Render(strings.Repeat("─", barW-filled))

	b.WriteString(fmt.Sprintf("  %s %s %s %s\n\n",
		m.spinner.View(),
		boldWhite.Render(fmt.Sprintf("%d/%d", checked, total)),
		bar,
		dimStyle.Render(fmt.Sprintf("%d%%", pct)),
	))

	// Stats row
	m.mu.Lock()
	avail := len(m.available)
	m.mu.Unlock()

	errors := int(stats.Errors)
	taken := checked - avail - errors

	statsLine := fmt.Sprintf("  %s %s  %s %s  %s %s  %s %s  %s %s  %s %s",
		dimStyle.Render("available"), greenBold.Render(fmt.Sprintf("%d", avail)),
		dimStyle.Render("taken"), boldWhite.Render(fmt.Sprintf("%d", taken)),
		dimStyle.Render("errors"), redStyle.Render(fmt.Sprintf("%d", errors)),
		dimStyle.Render("rate"), amberStyle.Render(fmt.Sprintf("%.0f/s", rate)),
		dimStyle.Render("elapsed"), cyanStyle.Render(elapsed.Round(time.Second).String()),
		dimStyle.Render("eta"), boldWhite.Render(eta),
	)
	b.WriteString(statsLine + "\n\n")

	// Available domains (max 8 lines)
	m.mu.Lock()
	availDomains := make([]string, len(m.available))
	copy(availDomains, m.available)
	logs := make([]string, len(m.logs))
	copy(logs, m.logs)
	m.mu.Unlock()

	if len(availDomains) > 0 {
		var av strings.Builder
		av.WriteString(greenBold.Render("● Available") + "\n\n")
		show := availDomains
		if len(show) > 8 {
			show = show[len(show)-8:]
		}
		for _, d := range show {
			av.WriteString(fmt.Sprintf("  %s %s\n", greenBold.Render("✓"), boldWhite.Render(d)))
		}
		if len(availDomains) > 8 {
			av.WriteString(dimStyle.Render(fmt.Sprintf("\n  + %d more (see results file)", len(availDomains)-8)))
		}
		b.WriteString(availBox.Render(av.String()) + "\n\n")
	}

	// Activity log (max 8 lines)
	var logContent strings.Builder
	logContent.WriteString(dimStyle.Bold(true).Render("● Activity") + "\n\n")
	start := 0
	if len(logs) > 8 {
		start = len(logs) - 8
	}
	for _, l := range logs[start:] {
		if strings.Contains(l, "✓") {
			logContent.WriteString("  " + greenBold.Render(l) + "\n")
		} else if strings.Contains(l, "⚠") {
			logContent.WriteString("  " + amberStyle.Render(l) + "\n")
		} else {
			logContent.WriteString("  " + dimStyle.Render(l) + "\n")
		}
	}
	b.WriteString(boxStyle.Render(logContent.String()) + "\n\n")

	// Footer
	b.WriteString(dimStyle.Render("  q quit • results saved in real-time"))

	return b.String()
}
