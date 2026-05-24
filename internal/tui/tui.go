package tui

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/codeatlasdev/domain-hunter/internal/pricing"
	"github.com/codeatlasdev/domain-hunter/internal/scanner"
)

var (
	blue      = lipgloss.Color("#2563EB")
	blueLight = lipgloss.Color("#60A5FA")
	green     = lipgloss.Color("#10B981")
	red       = lipgloss.Color("#EF4444")
	amber     = lipgloss.Color("#F59E0B")
	cyan      = lipgloss.Color("#06B6D4")
	purple    = lipgloss.Color("#A78BFA")
	slate300  = lipgloss.Color("#CBD5E1")
	slate400  = lipgloss.Color("#94A3B8")
	slate500  = lipgloss.Color("#64748B")
	slate600  = lipgloss.Color("#475569")
	slate700  = lipgloss.Color("#334155")
	slate800  = lipgloss.Color("#1E293B")
	slate900  = lipgloss.Color("#0F172A")
	white     = lipgloss.Color("#F8FAFC")
)

type availDomain struct {
	Domain  string
	Pricing *pricing.PriceResult
}

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
	available []availDomain
	logs      []string
	mu        sync.Mutex
	done      bool
	width     int
	height    int
}

type tickMsg struct{}
type doneMsg struct{}

func NewModel(sc *scanner.Scanner, cfg Config) *Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(blueLight)

	m := &Model{
		spinner: s,
		scanner: sc,
		config:  cfg,
		width:   100,
		height:  30,
	}

	sc.OnResult = func(r scanner.Result) {
		m.mu.Lock()
		defer m.mu.Unlock()
		ts := time.Now().Format("15:04:05")
		if r.Available {
			pr := pricing.GetPrices(r.Domain)
			m.available = append(m.available, availDomain{Domain: r.Domain, Pricing: &pr})
			m.logs = append(m.logs, fmt.Sprintf("[%s] ✓ %s", ts, r.Domain))
		} else if r.Error {
			m.logs = append(m.logs, fmt.Sprintf("[%s] ⚠ %s", ts, r.Domain))
		} else {
			m.logs = append(m.logs, fmt.Sprintf("[%s] · %s", ts, r.Domain))
		}
		if len(m.logs) > 500 {
			m.logs = m.logs[len(m.logs)-500:]
		}
	}

	return m
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, doTick(), waitDone(m.scanner.Done))
}

func doTick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg { return tickMsg{} })
}

func waitDone(ch chan struct{}) tea.Cmd {
	return func() tea.Msg {
		<-ch
		return doneMsg{}
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
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
	w := m.width
	h := m.height
	if w < 50 {
		w = 50
	}
	if h < 15 {
		h = 15
	}

	stats := m.scanner.Stats()
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
	eta := "—"
	if rate > 0 && checked < total {
		remaining := float64(total-checked) / rate
		eta = (time.Duration(remaining) * time.Second).Round(time.Second).String()
	}

	m.mu.Lock()
	availDomains := make([]availDomain, len(m.available))
	copy(availDomains, m.available)
	logs := make([]string, len(m.logs))
	copy(logs, m.logs)
	avail := len(m.available)
	m.mu.Unlock()

	errors := int(stats.Errors)
	taken := checked - avail - errors

	// ─── HEADER BAR ─────────────────────────────────────────────────────
	brand := lipgloss.NewStyle().Bold(true).Foreground(blueLight).Render("◆ domh")
	info := lipgloss.NewStyle().Foreground(slate400).Render(
		fmt.Sprintf("%d-letter .%s · %s · %d workers",
			m.config.Length, strings.Join(m.config.TLDs, ",."), m.config.Pattern, m.config.Workers))
	headerLeft := brand + "  " + info
	headerRight := lipgloss.NewStyle().Foreground(slate500).Render("q quit")
	headerGap := w - lipgloss.Width(headerLeft) - lipgloss.Width(headerRight) - 2
	if headerGap < 1 {
		headerGap = 1
	}
	header := " " + headerLeft + strings.Repeat(" ", headerGap) + headerRight + " "

	// ─── PROGRESS BAR ───────────────────────────────────────────────────
	barW := w - 30
	if barW < 15 {
		barW = 15
	}
	filled := (pct * barW) / 100
	barFilled := lipgloss.NewStyle().Foreground(blue).Bold(true).Render(strings.Repeat("█", filled))
	barEmpty := lipgloss.NewStyle().Foreground(slate700).Render(strings.Repeat("░", barW-filled))
	progressLeft := fmt.Sprintf(" %s %s",
		m.spinner.View(),
		lipgloss.NewStyle().Bold(true).Foreground(white).Render(fmt.Sprintf("%d/%d", checked, total)))
	progressRight := lipgloss.NewStyle().Foreground(blueLight).Bold(true).Render(fmt.Sprintf(" %d%%", pct))
	progress := progressLeft + " " + barFilled + barEmpty + progressRight

	// ─── STATS ROW ──────────────────────────────────────────────────────
	statsItems := []string{
		lipgloss.NewStyle().Foreground(green).Bold(true).Render(fmt.Sprintf("● %d available", avail)),
		lipgloss.NewStyle().Foreground(slate400).Render(fmt.Sprintf("○ %d taken", taken)),
		lipgloss.NewStyle().Foreground(amber).Render(fmt.Sprintf("⚠ %d errors", errors)),
		lipgloss.NewStyle().Foreground(cyan).Render(fmt.Sprintf("⚡ %.0f/s", rate)),
		lipgloss.NewStyle().Foreground(purple).Render(fmt.Sprintf("⏱ %s", elapsed.Round(time.Second))),
		lipgloss.NewStyle().Foreground(slate300).Render(fmt.Sprintf("→ %s", eta)),
	}
	statsRow := " " + strings.Join(statsItems, "  ")

	// ─── SEPARATOR ──────────────────────────────────────────────────────
	sep := lipgloss.NewStyle().Foreground(slate700).Render(strings.Repeat("─", w))

	// ─── PANELS (split layout) ──────────────────────────────────────────
	panelHeight := h - 7 // header + progress + stats + sep + footer
	if panelHeight < 5 {
		panelHeight = 5
	}

	leftW := w / 2
	rightW := w - leftW - 1 // -1 for separator

	// Left panel: Available domains
	leftContent := m.renderAvailable(availDomains, leftW-4, panelHeight-3)
	leftPanel := lipgloss.NewStyle().
		Width(leftW).
		Height(panelHeight).
		BorderRight(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(slate700).
		Padding(0, 1).
		Render(leftContent)

	// Right panel: Activity log
	rightContent := m.renderLog(logs, rightW-4, panelHeight-3)
	rightPanel := lipgloss.NewStyle().
		Width(rightW).
		Height(panelHeight).
		Padding(0, 1).
		Render(rightContent)

	panels := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)

	// ─── FOOTER ─────────────────────────────────────────────────────────
	footerLeft := lipgloss.NewStyle().Foreground(slate500).Render(" results auto-saved")
	footerRight := lipgloss.NewStyle().Foreground(slate600).Render("domh v1.4.0 ")
	footerGap := w - lipgloss.Width(footerLeft) - lipgloss.Width(footerRight)
	if footerGap < 1 {
		footerGap = 1
	}
	footer := footerLeft + strings.Repeat(" ", footerGap) + footerRight

	// ─── COMPOSE ────────────────────────────────────────────────────────
	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		progress,
		statsRow,
		sep,
		panels,
		sep,
		footer,
	)
}

func (m *Model) renderAvailable(domains []availDomain, maxW, maxLines int) string {
	var b strings.Builder
	title := lipgloss.NewStyle().Bold(true).Foreground(green).Render("✓ Available")
	count := lipgloss.NewStyle().Foreground(slate400).Render(fmt.Sprintf(" (%d)", len(domains)))
	b.WriteString(title + count + "\n\n")

	if len(domains) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(slate500).Italic(true).Render("  scanning..."))
		return b.String()
	}

	show := domains
	if len(show) > maxLines-2 {
		show = show[len(show)-(maxLines-2):]
	}

	for _, d := range show {
		domain := lipgloss.NewStyle().Bold(true).Foreground(white).Render(d.Domain)
		price := ""
		if d.Pricing != nil && d.Pricing.Cheapest != nil {
			price = lipgloss.NewStyle().Foreground(slate400).Render(
				fmt.Sprintf(" $%.0f %s", d.Pricing.Cheapest.RegisterPrice, d.Pricing.Cheapest.Registrar))
		}
		line := "  " + domain + price
		if lipgloss.Width(line) > maxW {
			line = "  " + domain
		}
		b.WriteString(line + "\n")
	}

	if len(domains) > maxLines-2 {
		b.WriteString(lipgloss.NewStyle().Foreground(slate500).Render(
			fmt.Sprintf("\n  + %d more", len(domains)-(maxLines-2))))
	}

	return b.String()
}

func (m *Model) renderLog(logs []string, maxW, maxLines int) string {
	var b strings.Builder
	title := lipgloss.NewStyle().Bold(true).Foreground(slate400).Render("◌ Activity")
	b.WriteString(title + "\n\n")

	if len(logs) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(slate500).Italic(true).Render("  waiting..."))
		return b.String()
	}

	start := 0
	if len(logs) > maxLines-2 {
		start = len(logs) - (maxLines - 2)
	}

	for _, l := range logs[start:] {
		if len(l) > maxW {
			l = l[:maxW-3] + "..."
		}
		if strings.Contains(l, "✓") {
			b.WriteString("  " + lipgloss.NewStyle().Foreground(green).Render(l) + "\n")
		} else if strings.Contains(l, "⚠") {
			b.WriteString("  " + lipgloss.NewStyle().Foreground(amber).Render(l) + "\n")
		} else {
			b.WriteString("  " + lipgloss.NewStyle().Foreground(slate600).Render(l) + "\n")
		}
	}

	return b.String()
}
