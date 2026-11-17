package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// shell composes the full-screen frame: header, body, status and key bar.
// Pages render only their body content; the shell owns the chrome so the
// layout is decided in one place instead of hand-counted in View().

const (
	headerRows = 1
	statusRows = 1
	footerRows = 1
)

func (m model) renderShell(body, footer string) string {
	header := m.renderHeader()
	status := m.statusLine()

	// the status row is only reserved when it says something: an idle magpie
	// spent a whole line on nothing
	rows := m.statusRowsUsed()
	bodyH := m.h - headerRows - footerRows - rows
	if bodyH < 1 {
		bodyH = 1
	}
	body = fitHeight(body, bodyH)

	// a two-line footer (the group detail page) is kept whole: it takes one
	// line from the body rather than being clipped
	if strings.Count(footer, "\n") >= 1 && bodyH > 1 {
		body = fitHeight(m.bodyForMode(), bodyH-1)
	}

	lines := []string{header, body}
	if status != "" {
		lines = append(lines, pad+status)
	}
	lines = append(lines, pad+footer)
	// a line wider than the window wraps, and the wrapped part is drawn over
	// the row below it: this frame is what the terminal gets, so nothing in it
	// is wider than the terminal
	return clipLines(strings.Join(lines, "\n"), m.w)
}

// clipLines cuts every line to the window's width. The footer's key bar and a
// long empty-state sentence reach past a narrow terminal, and one that wraps
// costs a row of the frame. A cut line ends in "…", so a key bar that goes on
// past the window reads as continuing rather than as ending there.
func clipLines(s string, w int) string {
	if w <= 0 {
		return s
	}
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		if lipgloss.Width(l) > w {
			if w <= 1 {
				lines[i] = "…"
				continue
			}
			lines[i] = lipgloss.NewStyle().MaxWidth(w-1).Render(l) + "…"
		}
	}
	return strings.Join(lines, "\n")
}

// bodyForMode returns the page/overlay body for the current mode, without
// the header — the shell renders the header once.
func (m model) bodyForMode() string {
	switch m.mode {
	case modeList:
		switch m.page {
		case pageProviders:
			return m.viewProvidersBody()
		case pageGroups:
			return m.viewGroupsBody()
		case pageLibrary:
			return m.viewLibraryBody()
		case pageSessions:
			return m.viewSessionsBody()
		case pageUsage:
			return m.viewUsageBody()
		default:
			return m.viewListBody()
		}
	case modePick, modeProfiles:
		return m.viewPickerBody()
	case modeAsk:
		return m.viewAskBody()
	case modeGroup:
		return m.viewGroupBody()
	case modeName:
		return m.viewNameBody()
	}
	return ""
}

// footerForMode returns the key bar for the current mode/page. The strings
// are literal hints(...) calls so tuiWords' AST scan still sees every word
// (i18n strategy A).
func (m model) footerForMode() string {
	switch m.mode {
	case modeList:
		switch m.page {
		case pageProviders:
			return hints("↑↓", "provider", "↵", "models", "e", "key", "w", "address", "a", "add", "f", "family", "u", "list/groups only", "o", "on/off", "m", "fetch models", "t", "test", "b", "balances", "d", "remove", "1–6", "pages")
		case pageGroups:
			return hints("↑↓", "group", "↵", "open", "n", "new", "o", "routing", "d", "remove", "u", "bring back", "1–6", "pages", "q", "quit")
		case pageLibrary:
			return hints("↑↓", "item", "↵", "agents", "a", "add", "e", "edit", "u", "update skill", "i", "bring in", "d", "remove", "s", "sync", "r", "reload", "1–6", "pages", "q", "quit")
		case pageSessions:
			if m.sstat {
				return hints("←→", "range", "M", "model", "f", "folder", "x", "clear", "c", "tokens / cost", "s", "sessions", "r", "reload", "1–6", "pages", "q", "quit")
			}
			return hints("↑↓", "session", "↵", "resume", "←→", "range", "M", "model", "f", "folder", "x", "clear", "s", "stats", "r", "reload", "1–6", "pages", "q", "quit")
		case pageUsage:
			return hints("←→", "period", "t w m A", "today · 7 days · 30 days · all", "u", "used / left", "c", "daily check-in", "r", "reload", "1–6", "pages", "q", "quit")
		default:
			return listFooter()
		}
	case modePick, modeProfiles:
		switch {
		case m.mode == modeProfiles:
			return hints("↑↓", "move", "↵", "apply", "ctrl+e", "rename", "ctrl+d", "delete", "esc", "back")
		case m.pk.refetch != nil:
			return hints("↑↓", "move", "↵", "on / off", "ctrl+r", "fetch models again", "esc", "back")
		case m.pk.toggle != nil:
			return hints("↑↓", "move", "↵", "on / off", "esc", "back")
		default:
			return hints("↑↓", "move", "↵", "select", "esc", "back")
		}
	case modeAsk:
		return hints("↵", "save", "esc", "cancel")
	case modeGroup:
		return hints("↑↓", "model / rule", "J K", "move", "a", "add model", "n", "new rule", "↵", "edit rule", "d", "take out", "esc", "back") + "\n" +
			pad + hints("E", "effort", "F", "fast", "c", "classifier", "o", "routing", "s", "stays", "x", "context", "l", "levels", "f", "family", "N", "note", "R", "rename")
	case modeName:
		if m.rename != "" {
			return hints("↵", "rename", "esc", "cancel")
		}
		return hints("↵", "save", "esc", "cancel")
	}
	return ""
}

func (m model) statusLine() string {
	switch {
	case m.flash != "" && m.flashOK:
		return sOK.Render("✓ ") + sText.Render(m.flash)
	case m.flash != "":
		return sBad.Render("✗ ") + sText.Render(m.flash)
	case m.syncing:
		return sMuted.Render("… " + tr("syncing model catalog"))
	default:
		return gatewayStatus()
	}
}

// statusRowsUsed is how many rows the status line takes right now: an idle
// magpie says nothing there.
func (m model) statusRowsUsed() int {
	if m.statusLine() == "" {
		return 0
	}
	return statusRows
}

// renderHeader draws the global top bar: brand, page tabs, and a right-aligned
// gateway badge. It is the same bar in every mode, so an overlay keeps the
// page it was opened from in view; the overlay's own crumbs are in its body.
func (m model) renderHeader() string {
	return m.renderTabHeader()
}

func (m model) renderTabHeader() string {
	// the tabs give way in steps as the terminal narrows, never by wrapping:
	// a header wider than the window shifts every row below it down one
	brand := pad + sTitle.Render("◉ magpie") + "  "
	left := brand + m.tabStrip(false)
	if m.w > 0 && lipgloss.Width(left) > m.w {
		for _, cand := range []string{
			m.tabStrip(false),    // the six tabs, brand dropped
			brand + m.pageChip(), // the page in hand, named, alone
			pad + m.pageChip(),
			brand + m.tabStrip(true), // each tab at one letter
			m.tabStrip(true),
		} {
			left = cand
			if lipgloss.Width(cand) <= m.w {
				break
			}
		}
		if m.w > 0 && lipgloss.Width(left) > m.w {
			left = trunc(left, m.w)
		}
	}

	right := m.headerBadge()
	if right == "" {
		return left
	}
	gap := m.w - lipgloss.Width(left) - lipgloss.Width(right) - 1
	if gap < 2 {
		return left
	}
	return left + strings.Repeat(" ", gap) + right
}

// tabStrip draws the six page tabs, shortened to one letter each when short.
func (m model) tabStrip(short bool) string {
	var b strings.Builder
	for i, n := range pageNames {
		label := fmt.Sprintf("%d %s", i+1, tr(n))
		if short {
			label = fmt.Sprintf("%d%s", i+1, firstLetter(tr(n)))
		}
		if page(i) == m.page {
			b.WriteString(sTabOn.Render(label) + " ")
		} else {
			b.WriteString(sTabOff.Render(" "+label+" ") + " ")
		}
	}
	return strings.TrimRight(b.String(), " ")
}

func (m model) pageTabLabel() string {
	return fmt.Sprintf("%d %s", int(m.page)+1, tr(pageNames[m.page]))
}

// pageChip is the current page alone with its place among the pages, for a
// terminal with no room for the tabs.
func (m model) pageChip() string {
	return sTabOn.Render(m.pageTabLabel()) + " " +
		sMuted.Render(fmt.Sprintf("%d/%d", int(m.page)+1, len(pageNames)))
}

// firstLetter is the tab's own letter, shortened for a narrow terminal: runes,
// so a two-byte letter is not cut in half.
func firstLetter(s string) string {
	for _, r := range s {
		return string(r)
	}
	return s
}

func (m model) headerBadge() string {
	if gw.here.Load() {
		return sOK.Render("●") + " " + sMuted.Render(tr("gateway"))
	}
	if e := gw.err.Load(); e != nil && *e != "" {
		return sBad.Render("●") + " " + sMuted.Render(tr("gateway"))
	}
	return ""
}

// crumbHeader is the breadcrumb form used inside overlay bodies.
func (m model) crumbHeader(crumbs ...string) string {
	s := pad + sTitle.Render("◉ magpie")
	for _, c := range crumbs {
		s += sCrumb.Render(" › ") + sText.Render(c)
	}
	return s
}

func fitHeight(s string, h int) string {
	if h <= 0 {
		return ""
	}
	lines := strings.Split(s, "\n")
	if len(lines) > h {
		lines = lines[:h]
	}
	if len(lines) < h {
		lines = append(lines, make([]string, h-len(lines))...)
	}
	return strings.Join(lines, "\n")
}
