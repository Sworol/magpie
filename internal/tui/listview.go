package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// listFooter is the Agents page's footer. Its words are written out here and
// not read from a table of the keys, though one would say the same: the i18n
// test reads every word the source hands tr and hints (tuiWords, i18n_test.go)
// and fails on a zh word the TUI no longer shows, so a word living only in a
// table is a word it cannot see. A table-driven footer wants that reader taught
// about the table first.
func listFooter() string {
	return hints(
		"↑↓", "agent",
		"←→", "field",
		"↵", "change",
		"s", "save profile",
		"p", "profiles",
		"1–6", "pages",
		"q", "quit",
	)
}

const detailStripRows = 4

type agentItem struct {
	idx    int
	name   string
	desc   string
	filter string
	hidden bool
}

func (i agentItem) Title() string       { return i.name }
func (i agentItem) Description() string { return i.desc }
func (i agentItem) FilterValue() string { return i.filter }

type agentDelegate struct{}

func (agentDelegate) Height() int                             { return 1 }
func (agentDelegate) Spacing() int                            { return 0 }
func (agentDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (agentDelegate) Render(w io.Writer, m list.Model, index int, it list.Item) {
	a, ok := it.(agentItem)
	if !ok {
		return
	}
	selected := index == m.Index()
	name := a.name
	if a.hidden {
		name = sFaint.Render(name)
	}
	// the description starts in the same column on every row, so a long agent
	// name pushes only its own row and the values stay scannable
	width := nameColumn(m.Items()) - 2
	if selected {
		// "▸ <name>", the marker the page's tests read for the row picked
		fmt.Fprintf(w, "%s%s%s", pad, sCursor.Render("▸ "), sNameOn.Render(padRight(name, width)))
	} else {
		fmt.Fprintf(w, "%s  %s", pad, sText.Render(padRight(name, width)))
	}
	desc := a.desc
	if desc != "" {
		if selected {
			fmt.Fprintf(w, "  %s", sText.Render(desc))
		} else {
			fmt.Fprintf(w, "  %s", sMuted.Render(desc))
		}
	}
}

// nameColumn is the width the agent names are padded to: the widest of them,
// so every description lines up.
func nameColumn(items []list.Item) int {
	w := 0
	for _, it := range items {
		if a, ok := it.(agentItem); ok {
			w = max(w, lipgloss.Width(a.name))
		}
	}
	return w
}

func newAgentList(items []list.Item) list.Model {
	d := agentDelegate{}
	l := list.New(items, d, 0, 0)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)
	l.SetShowPagination(false)
	l.DisableQuitKeybindings()
	l.KeyMap.Quit.SetKeys()
	l.KeyMap.ForceQuit.SetKeys()
	l.FilterInput.Prompt = pad + sCursor.Render("❯ ")
	l.FilterInput.PromptStyle = sCursor
	l.FilterInput.Cursor.Style = sCursor
	l.Styles.NoItems = sMuted
	return l
}

func agentItems(m model) []list.Item {
	items := make([]list.Item, 0, len(m.agents))
	for i, a := range m.agents {
		var vals map[string]string
		if i < len(m.values) {
			vals = m.values[i]
		}
		descParts := make([]string, 0, 4)
		filterParts := []string{a.Name, a.ID}
		for _, f := range a.Fields {
			v := vals[f.Key]
			filterParts = append(filterParts, v)
			if v == "" {
				if f.Quiet {
					continue
				}
				continue
			}
			descParts = append(descParts, v)
		}
		fail := ""
		if i < len(m.failover) {
			fail = m.failover[i]
		}
		if fail != "" {
			descParts = append(descParts, fail)
		}
		if i >= m.hidden {
			descParts = append(descParts, tr("hidden"))
		}
		items = append(items, agentItem{
			idx:    i,
			name:   a.Name,
			desc:   strings.Join(descParts, "  ·  "),
			filter: strings.Join(filterParts, " "),
			hidden: i >= m.hidden,
		})
	}
	return items
}

// renderList is the list to draw. The model's own is used as it stands: it
// holds the cursor and, while / is typed into, the items the filter left, so
// setting them again here would draw every agent as if nothing had been typed.
// The cursor is only set for a model that never made a list of its own (a
// test's struct literal), where there is no state to keep.
func (m model) renderList() list.Model {
	l := m.alist
	if !m.alistReady {
		items := agentItems(m)
		l = newAgentList(items)
		if len(items) > 0 {
			l.Select(clamp(m.row, len(items)))
		}
	}
	l.SetSize(m.listWidth(), m.listRows())
	return l
}

// listRows is the list's own height: what the shell leaves the body, less the
// status row when it is there, the strip under the list and the blank line
// above it.
func (m model) listRows() int {
	h := m.h - headerRows - m.statusRowsUsed() - footerRows - detailStripRows - 1
	return max(3, h)
}

func (m model) listWidth() int { return max(20, m.w) }

// initAgentList makes the model's own list before it is used: the zero
// list.Model panics in SetItems on the paginator it never made.
func (m *model) initAgentList() {
	if m.alistReady {
		return
	}
	m.alist = newAgentList(nil)
	m.alistReady = true
}

func (m *model) syncAgentList() {
	m.initAgentList()
	items := agentItems(*m)
	m.alist.SetItems(items)
	m.alist.SetSize(m.listWidth(), m.listRows())
	if len(items) == 0 {
		m.row = 0
		return
	}
	m.row = clamp(m.row, len(items))
	m.alist.Select(m.row)
}

func (m model) currentAgentIndex() int {
	if len(m.agents) == 0 {
		return -1
	}
	if m.alistReady {
		if it, ok := m.alist.SelectedItem().(agentItem); ok {
			return it.idx
		}
	}
	return clamp(m.row, len(m.agents))
}

// viewList is the Agents page including header (used by tests and direct calls).
func (m model) viewList() string {
	if len(m.agents) == 0 {
		return m.header()
	}
	return m.header() + "\n\n" + m.viewListBody()
}

// viewListBody renders the bubbles list and the detail strip.
func (m model) viewListBody() string {
	if len(m.agents) == 0 {
		return pad + "  " + sMuted.Render(tr("no supported agents found on this machine"))
	}
	var b strings.Builder
	b.WriteString(m.renderList().View())
	b.WriteString("\n")
	b.WriteString(m.agentDetailStrip())
	return strings.TrimRight(b.String(), "\n")
}

func (m model) agentDetailStrip() string {
	i := m.currentAgentIndex()
	if i < 0 || i >= len(m.agents) {
		return sFaint.Render(strings.Repeat("─", max(4, m.w-len(pad)*2-2)))
	}
	a := m.agents[i]
	var vals map[string]string
	if i < len(m.values) {
		vals = m.values[i]
	}
	path := tilde(a.Path)
	ruleW := max(4, m.w-len(pad)*2-2)
	var out []string
	out = append(out, pad+sRule.Render(strings.Repeat("─", ruleW)))

	title := pad + sNameOn.Render(a.Name)
	if path != "" {
		gap := m.w - lipgloss.Width(title) - lipgloss.Width(path) - 2
		if gap > 1 {
			title += strings.Repeat(" ", gap) + sFaint.Render(path)
		} else {
			out = append(out, title)
			out = append(out, pad+sFaint.Render(path))
			title = ""
		}
	}
	if title != "" {
		out = append(out, title)
	}

	// fields line with selected field highlighted (←→ moves m.col)
	var cells []string
	for j, f := range a.Fields {
		v := ""
		if vals != nil {
			v = vals[f.Key]
		}
		if v == "" && f.Quiet && j != m.col {
			continue
		}
		label := sMuted.Render(tr(f.Label) + " ")
		val := dash(v)
		cell := label + sCell.Render(val)
		if j == m.col {
			cell = label + sCellOn.Render(padRight(val, max(1, lipgloss.Width(val))))
		}
		cells = append(cells, cell)
	}
	if len(cells) > 0 {
		out = append(out, pad+strings.Join(cells, sMuted.Render("  ·  ")))
	}
	fail := ""
	if i < len(m.failover) {
		fail = m.failover[i]
	}
	if fail != "" {
		out = append(out, pad+sWarn.Render(fail))
	}
	for len(out) < detailStripRows {
		out = append(out, "")
	}
	if len(out) > detailStripRows {
		out = out[:detailStripRows]
	}
	return strings.Join(out, "\n")
}
