package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palette — dark theme with vibrant accents.
var (
	ColorPrimary   = lipgloss.Color("#7C3AED") // violet
	ColorSecondary = lipgloss.Color("#06B6D4") // cyan
	ColorSuccess   = lipgloss.Color("#22C55E") // green
	ColorWarning   = lipgloss.Color("#F59E0B") // amber
	ColorDanger    = lipgloss.Color("#EF4444") // red
	ColorMuted     = lipgloss.Color("#6B7280") // gray
	ColorText      = lipgloss.Color("#F9FAFB") // near-white
	ColorDimText   = lipgloss.Color("#9CA3AF") // dim gray
	ColorBg        = lipgloss.Color("#111827") // dark bg
	ColorSidebarBg = lipgloss.Color("#1F2937") // sidebar bg
	ColorPanelBg   = lipgloss.Color("#0F172A") // panel bg
	ColorBorder    = lipgloss.Color("#374151") // subtle border
	ColorHighlight = lipgloss.Color("#312E81") // selected bg
)

// Sidebar styles.
var (
	SidebarStyle = lipgloss.NewStyle().
			Background(ColorSidebarBg).
			Padding(1, 1)

	SidebarTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorPrimary).
				MarginBottom(1)

	AgentItemStyle = lipgloss.NewStyle().
			Foreground(ColorText).
			PaddingLeft(1)

	AgentItemSelectedStyle = lipgloss.NewStyle().
				Foreground(ColorText).
				Background(ColorHighlight).
				Bold(true).
				PaddingLeft(1)

	AgentStatusRunning = lipgloss.NewStyle().
				Foreground(ColorSuccess)

	AgentStatusStopped = lipgloss.NewStyle().
				Foreground(ColorDanger)
)

// Log viewer styles.
var (
	LogPanelStyle = lipgloss.NewStyle().
			Background(ColorPanelBg).
			Padding(0, 1)

	LogTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorSecondary).
			MarginBottom(1)

	LogContentStyle = lipgloss.NewStyle().
			Foreground(ColorDimText)

	LogPlaceholderStyle = lipgloss.NewStyle().
				Foreground(ColorMuted).
				Italic(true).
				Align(lipgloss.Center)
)

// Status bar styles.
var (
	StatusBarStyle = lipgloss.NewStyle().
			Background(ColorPrimary).
			Foreground(ColorText).
			Padding(0, 1)

	StatusBarKeyStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorText)

	StatusBarDescStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#DDD6FE"))

	StatusBarInfoStyle = lipgloss.NewStyle().
				Foreground(ColorText).
				Bold(true)
)

// Border style for panels.
var (
	PanelBorderStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder)
)
