package tui

// The terminal's whole look: every colour and style a page draws with lives
// here, so changing one is this file's business and not four pages'.

import "github.com/charmbracelet/lipgloss"

// ---- styling ---------------------------------------------------------------

var (
	cAccent = lipgloss.AdaptiveColor{Light: "#4338CA", Dark: "#A5B4FC"}
	cText   = lipgloss.AdaptiveColor{Light: "#1F2328", Dark: "#E6E6E6"}
	cMuted  = lipgloss.AdaptiveColor{Light: "#6B7280", Dark: "#8B93A1"}
	cFaint  = lipgloss.AdaptiveColor{Light: "#AEB4BF", Dark: "#4A4F5A"}
	cOK     = lipgloss.AdaptiveColor{Light: "#0F9D58", Dark: "#7EE2A8"}
	cBad    = lipgloss.AdaptiveColor{Light: "#D93025", Dark: "#FF8A80"}
	cWarn   = lipgloss.AdaptiveColor{Light: "#B26A00", Dark: "#FFCA7A"}
	cPill   = lipgloss.AdaptiveColor{Light: "#EEF0FF", Dark: "#2A2C42"}
	cLine   = lipgloss.AdaptiveColor{Light: "#DFE2E8", Dark: "#33363F"}

	sTitle  = lipgloss.NewStyle().Bold(true).Foreground(cAccent)
	sCrumb  = lipgloss.NewStyle().Foreground(cMuted)
	sText   = lipgloss.NewStyle().Foreground(cText)
	sMuted  = lipgloss.NewStyle().Foreground(cMuted)
	sFaint  = lipgloss.NewStyle().Foreground(cFaint)
	sName   = lipgloss.NewStyle().Bold(true).Foreground(cText)
	sNameOn = lipgloss.NewStyle().Bold(true).Foreground(cAccent)
	sPill   = lipgloss.NewStyle().Foreground(cAccent).Background(cPill).Padding(0, 1)
	sValue  = lipgloss.NewStyle().Foreground(cText).Padding(0, 1)
	sOK     = lipgloss.NewStyle().Foreground(cOK)
	sBad    = lipgloss.NewStyle().Foreground(cBad)
	sWarn   = lipgloss.NewStyle().Foreground(cWarn)
	sKey    = lipgloss.NewStyle().Foreground(cText)
	sCursor = lipgloss.NewStyle().Foreground(cAccent).Bold(true)

	// tabs in the global header
	sTabOn  = lipgloss.NewStyle().Bold(true).Foreground(cAccent).Background(cPill).Padding(0, 1)
	sTabOff = lipgloss.NewStyle().Foreground(cMuted)

	// panels and detail strips
	sPanel      = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(cLine).Padding(0, 1)
	sPanelFocus = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(cAccent).Padding(0, 1)
	sRule       = lipgloss.NewStyle().Foreground(cLine)

	// the table's own, padding-free: a cell padded to its column's width must
	// keep the columns under one another, so none of these may add padding
	sHead   = lipgloss.NewStyle().Bold(true).Foreground(cMuted)
	sCell   = lipgloss.NewStyle().Foreground(cText)
	sCellOn = lipgloss.NewStyle().Bold(true).Foreground(cAccent).Background(cPill)
)

const pad = "  "
