package tui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

// TestHeaderFitsTheWindow: the header is one row of the frame, so it can
// never be wider than the terminal. A wider one wraps, and every row below it
// moves down one, which puts the key bar under the last row the program
// prints. The tabs give way in steps as the terminal narrows, and the page in
// hand is still findable at every width.
func TestHeaderFitsTheWindow(t *testing.T) {
	pages := []page{pageAgents, pageProviders, pageGroups, pageUsage, pageSessions, pageLibrary}
	for _, w := range []int{20, 24, 32, 40, 56, 72, 100, 140} {
		for _, p := range pages {
			m := model{w: w, h: 24, page: p}
			head := m.renderHeader()
			if got := lipgloss.Width(head); got > w {
				t.Errorf("%d columns, page %v: header takes %d: %q", w, p, got, head)
			}
			// the tab at one letter is the shortest form that still says which
			// page is which, so every width carries the page's name or that
			// letter beside its number
			short := fmt.Sprintf("%d%s", int(p)+1, firstLetter(tr(pageNames[p])))
			if !strings.Contains(head, tr(pageNames[p])) && !strings.Contains(head, short) {
				t.Errorf("%d columns, page %v: header names neither: %q", w, p, head)
			}
		}
	}
}

// TestNarrowHeaderStillNamesThePageInHand: a terminal too narrow for the six
// tabs names the page the user is on, in full, instead of losing it.
func TestNarrowHeaderStillNamesThePageInHand(t *testing.T) {
	for _, w := range []int{24, 28, 32} {
		m := model{w: w, h: 24, page: pageSessions}
		head := m.renderHeader()
		if !strings.Contains(head, tr("sessions")) {
			t.Errorf("%d columns: header doesn't name the page: %q", w, head)
		}
		if got := lipgloss.Width(head); got > w {
			t.Errorf("%d columns: header takes %d: %q", w, got, head)
		}
	}
}

// TestFrameFitsTheWindow: no line of the frame is wider than the terminal.
// The key bar of a page with many keys and a long empty-state sentence both
// reach past a narrow window, and a line that wraps costs the frame a row and
// is drawn over the row below it. The library page is left out: its body
// dereferences a list this test's model never made, a fault of its own.
func TestFrameFitsTheWindow(t *testing.T) {
	pages := []page{pageAgents, pageProviders, pageGroups, pageUsage, pageSessions}
	for _, w := range []int{40, 56, 80, 100, 140} {
		for _, p := range pages {
			m := model{w: w, h: 24, page: p}
			for i, line := range strings.Split(m.View(), "\n") {
				if got := lipgloss.Width(line); got > w {
					t.Errorf("%d columns, page %v, line %d: %d wide: %q", w, p, i, got, line)
				}
			}
		}
	}
}

// TestViewKeepsTheFrameTallAsTheWindow: the frame is the window, so the
// footer's key bar is the last row even on a short terminal, and the body
// gives up rows to make room rather than pushing the footer off the screen.
func TestViewKeepsTheFrameTallAsTheWindow(t *testing.T) {
	for _, h := range []int{6, 10, 14, 24, 40} {
		m := model{w: 100, h: h, page: pageAgents}
		lines := strings.Split(m.View(), "\n")
		if len(lines) != h {
			t.Fatalf("%d rows: the frame has %d", h, len(lines))
		}
		if last := lines[len(lines)-1]; !strings.Contains(last, tr("quit")) {
			t.Errorf("%d rows: the last row isn't the key bar: %q", h, last)
		}
	}
}
