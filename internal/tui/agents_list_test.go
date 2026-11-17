package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/yetone/magpie/internal/agent"
)

// agentsModel is a model of three agents with a list of its own, as newModel
// leaves it: the values read, and an item for each agent.
func agentsModel(t *testing.T) model {
	t.Helper()
	f := agent.Field{Key: "model", Label: "model"}
	m := model{w: 100, h: 20, hidden: 3, page: pageAgents}
	for _, a := range []*agent.Agent{
		{ID: "claude", Name: "Claude Code", Fields: []agent.Field{f}},
		{ID: "codex", Name: "Codex", Fields: []agent.Field{f}},
		{ID: "opencode", Name: "OpenCode", Fields: []agent.Field{f}},
	} {
		m.agents = append(m.agents, a)
		m.values = append(m.values, map[string]string{"model": a.ID + "/m1"})
		m.failover = append(m.failover, "")
	}
	m.initAgentList()
	m.syncAgentList()
	return m
}

// typeRunes types a filter a key at a time, as a terminal sends it, and runs
// what the list asks for. A list leaves a filter alone when a whole word
// arrives in one key, and it narrows the list from a command of its own
// (FilterMatchesMsg) returned inside a tea.Batch, which the runtime unpacks
// and a test has to unpack itself.
func typeRunes(t *testing.T, m model, s string) model {
	t.Helper()
	for _, r := range s {
		next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = next.(model)
		m = runBatch(t, m, cmd, 0)
	}
	return m
}

func runBatch(t *testing.T, m model, cmd tea.Cmd, depth int) model {
	t.Helper()
	if cmd == nil || depth > 6 {
		return m
	}
	switch msg := cmd().(type) {
	case nil:
		return m
	case tea.BatchMsg:
		for _, c := range msg {
			m = runBatch(t, m, c, depth+1)
		}
		return m
	default:
		next, c := m.Update(msg)
		return runBatch(t, next.(model), c, depth+1)
	}
}

// TestAgentsFilterNarrowsTheList: / takes a word and the list keeps only the
// agents it matches, which is how a machine with more agents than rows finds
// one. The cursor follows the filter, so the strip below and the fields it
// edits are the agent shown. Enter leaves the filter applied.
func TestAgentsFilterNarrowsTheList(t *testing.T) {
	m := agentsModel(t)

	m = press(t, m, "/")
	if got := m.alist.FilterState(); got != list.Filtering {
		t.Fatalf("/ left the list %v, want filtering", got)
	}
	m = typeRunes(t, m, "codex")

	view := m.View()
	if !strings.Contains(view, "Codex") {
		t.Error("the agent typed isn't shown")
	}
	for _, gone := range []string{"Claude Code", "OpenCode"} {
		if strings.Contains(view, gone) {
			t.Errorf("%q is still shown with \"codex\" typed", gone)
		}
	}
	if i := m.currentAgentIndex(); i != 1 {
		t.Errorf("the cursor is on agent %d, want codex at 1", i)
	}

	m = press(t, m, "enter")
	if got := m.alist.FilterState(); got == list.Filtering {
		t.Error("enter left the filter open")
	}
	if !strings.Contains(m.View(), "Codex") || strings.Contains(m.View(), "OpenCode") {
		t.Error("the filter stopped narrowing the list after enter")
	}

	// / again opens the filter with what was typed, so a word can be changed
	m = press(t, m, "/")
	if got := m.alist.FilterState(); got != list.Filtering {
		t.Errorf("/ over an applied filter left the list %v", got)
	}
}

// TestAgentsFilterWithoutAgentsDoesNotTrap: a machine with no agent has
// nothing to filter, and the list refuses to start one rather than leaving
// the user typing into a filter that can never match.
func TestAgentsFilterWithoutAgentsDoesNotTrap(t *testing.T) {
	m := model{w: 100, h: 20, page: pageAgents}
	m.initAgentList()
	m.syncAgentList()
	m = press(t, m, "/")
	if got := m.alist.FilterState(); got == list.Filtering {
		t.Errorf("an empty list started filtering: %v", got)
	}
}
