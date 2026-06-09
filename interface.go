package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ── styles (Catppuccin Mocha palette) ────────────────────────────────────────

var (
	styleHeader  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#CBA6F7"))
	styleSubtle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086"))
	styleRule    = lipgloss.NewStyle().Foreground(lipgloss.Color("#313244"))
	styleName    = lipgloss.NewStyle().Foreground(lipgloss.Color("#CDD6F4"))
	styleVersion = lipgloss.NewStyle().Foreground(lipgloss.Color("#585B70"))
	styleOk      = lipgloss.NewStyle().Foreground(lipgloss.Color("#A6E3A1"))
	styleUpd     = lipgloss.NewStyle().Foreground(lipgloss.Color("#F9E2AF"))
	styleMiss    = lipgloss.NewStyle().Foreground(lipgloss.Color("#F38BA8"))
)

// ── state types ──────────────────────────────────────────────────────────────

type pkgStatus int

const (
	statusPending pkgStatus = iota
	statusOk
	statusOutdated
	statusNotFound
)

type pkgEntry struct {
	name          string
	localVersion  string
	remoteVersion string
	status        pkgStatus
}

type appState int

const (
	stateLoadingLocal appState = iota
	stateQueryingAUR
	stateRevealing
	stateDone
)

// ── model ────────────────────────────────────────────────────────────────────

type model struct {
	state    appState
	locals   []pac
	packages []pkgEntry
	spinner  spinner.Model
	err      error
	maxLen   int
	revealed int
}

// ── messages ─────────────────────────────────────────────────────────────────

type (
	localPackagesMsg []pac
	aurResponseMsg   *responseAur
	pkgRevealMsg     int
	errMsg           error
)

// ── constructor ──────────────────────────────────────────────────────────────

func initialModel() model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#CBA6F7"))
	return model{state: stateLoadingLocal, spinner: s}
}

// ── init ─────────────────────────────────────────────────────────────────────

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, fetchLocalCmd())
}

// ── commands ─────────────────────────────────────────────────────────────────

func fetchLocalCmd() tea.Cmd {
	return func() tea.Msg {
		raw, err := getPac()
		if err != nil {
			return errMsg(err)
		}
		return localPackagesMsg(parsePac(raw))
	}
}

func queryAurCmd(pacs []pac) tea.Cmd {
	return func() tea.Msg {
		resp, err := checkAur(pacs)
		if err != nil {
			return errMsg(err)
		}
		return aurResponseMsg(resp)
	}
}

// revealAfterDelay fires a pkgRevealMsg for the given index after a random 1–5s delay.
// The delay is computed eagerly so all goroutines start their timers concurrently.
func revealAfterDelay(index int) tea.Cmd {
	delay := time.Duration(rand.Intn(5)+1) * time.Second
	return func() tea.Msg {
		time.Sleep(delay)
		return pkgRevealMsg(index)
	}
}

// ── update ───────────────────────────────────────────────────────────────────

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case errMsg:
		m.err = msg
		return m, tea.Quit

	case spinner.TickMsg:
		if m.state == stateDone {
			return m, nil
		}
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case localPackagesMsg:
		m.locals = []pac(msg)
		if len(m.locals) == 0 {
			m.state = stateDone
			return m, tea.Quit
		}
		for _, p := range m.locals {
			if len(p.name) > m.maxLen {
				m.maxLen = len(p.name)
			}
		}
		m.state = stateQueryingAUR
		return m, queryAurCmd(m.locals)

	case aurResponseMsg:
		resp := (*responseAur)(msg)

		// index remote results by package name for O(1) lookup
		versionMap := make(map[string]string)
		for _, rem := range resp.Results {
			versionMap[rem.Name] = rem.Version
		}

		// build the display list and fire one reveal goroutine per package
		m.packages = make([]pkgEntry, len(m.locals))
		cmds := make([]tea.Cmd, len(m.locals))
		for i, loc := range m.locals {
			m.packages[i] = pkgEntry{
				name:          loc.name,
				localVersion:  loc.version,
				remoteVersion: versionMap[loc.name],
				status:        statusPending,
			}
			cmds[i] = revealAfterDelay(i)
		}
		m.state = stateRevealing
		return m, tea.Batch(cmds...)

	case pkgRevealMsg:
		i := int(msg)
		e := &m.packages[i]
		switch {
		case e.remoteVersion == "":
			e.status = statusNotFound
		case e.remoteVersion != e.localVersion:
			e.status = statusOutdated
		default:
			e.status = statusOk
		}
		m.revealed++
		if m.revealed == len(m.packages) {
			m.state = stateDone
			return m, tea.Quit
		}
	}

	return m, nil
}

// ── view ─────────────────────────────────────────────────────────────────────

const ruleWidth = 46

func (m model) View() string {
	if m.err != nil {
		return "\n  " + styleMiss.Render("✗  "+m.err.Error()) + "\n\n"
	}

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString("  " + styleHeader.Render("🔍 gopacgo") + "  " + styleSubtle.Render("AUR package checker") + "\n")
	b.WriteString("  " + styleRule.Render(strings.Repeat("─", ruleWidth)) + "\n\n")

	switch m.state {
	case stateLoadingLocal:
		b.WriteString("  " + m.spinner.View() + "  " + styleSubtle.Render("fetching local packages...") + "\n")

	case stateQueryingAUR:
		b.WriteString("  " + m.spinner.View() + "  " + styleSubtle.Render("querying AUR...") + "\n")

	case stateRevealing, stateDone:
		for _, pkg := range m.packages {
			b.WriteString(renderPkg(pkg, m.maxLen, m.spinner.View()))
			b.WriteString("\n")
		}
		b.WriteString("\n")
		b.WriteString("  " + styleRule.Render(strings.Repeat("─", ruleWidth)) + "\n")
		b.WriteString(renderSummary(m.packages) + "\n")
	}

	b.WriteString("\n")
	return b.String()
}

func renderPkg(e pkgEntry, maxLen int, spin string) string {
	name := styleName.Render(fmt.Sprintf("  %-*s", maxLen, e.name))
	ver := styleVersion.Render(fmt.Sprintf("  %-24s", e.localVersion))

	var status string
	switch e.status {
	case statusPending:
		status = styleSubtle.Render(spin + "  checking")
	case statusOk:
		status = styleOk.Render("✓  up to date")
	case statusOutdated:
		status = styleUpd.Render("↑  " + e.remoteVersion)
	case statusNotFound:
		status = styleMiss.Render("✗  not found")
	}

	return name + ver + status
}

func renderSummary(pkgs []pkgEntry) string {
	var cntOk, cntUpd, cntMiss, cntPend int
	for _, p := range pkgs {
		switch p.status {
		case statusOk:
			cntOk++
		case statusOutdated:
			cntUpd++
		case statusNotFound:
			cntMiss++
		case statusPending:
			cntPend++
		}
	}

	dot := styleSubtle.Render("  ·  ")
	var parts []string
	if cntOk > 0 {
		parts = append(parts, styleOk.Render(fmt.Sprintf("✓  %d up to date", cntOk)))
	}
	if cntUpd > 0 {
		parts = append(parts, styleUpd.Render(fmt.Sprintf("↑  %d to update", cntUpd)))
	}
	if cntMiss > 0 {
		parts = append(parts, styleMiss.Render(fmt.Sprintf("✗  %d not found", cntMiss)))
	}
	if cntPend > 0 {
		parts = append(parts, styleSubtle.Render(fmt.Sprintf("·  %d checking", cntPend)))
	}

	return "  " + strings.Join(parts, dot)
}
