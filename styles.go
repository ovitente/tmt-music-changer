package main

import "github.com/charmbracelet/lipgloss"

var (
	borderColor       = lipgloss.Color("#3a3a3a")
	activeBorderColor = lipgloss.Color("#606060")
	cursorColor       = lipgloss.Color("99")
	checkedColor      = lipgloss.Color("78")
	dirtyColor        = lipgloss.Color("203")
	commentColor      = lipgloss.Color("241")
	statusColor       = lipgloss.Color("241")
	reorderMoveColor  = lipgloss.Color("214")

	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(0, 1)

	activePanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(activeBorderColor).
				Padding(0, 1)

	cursorStyle = lipgloss.NewStyle().
			Foreground(cursorColor).
			Bold(true)

	normalItemStyle = lipgloss.NewStyle()

	checkedStyle = lipgloss.NewStyle().
			Foreground(checkedColor)

	dirtyStyle = lipgloss.NewStyle().
			Foreground(dirtyColor).
			Bold(true)

	commentStyle = lipgloss.NewStyle().
			Foreground(commentColor).
			Italic(true)

	statusStyle = lipgloss.NewStyle().
			Foreground(statusColor)

	titleStyle = lipgloss.NewStyle().
			Bold(true)

	reorderHighlight = lipgloss.NewStyle().
				Foreground(reorderMoveColor).
				Bold(true)

	searchInputStyle = lipgloss.NewStyle().
				Foreground(reorderMoveColor).
				Bold(true)

	searchHighlightStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("214")).
				Bold(true)

	helpKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true)

	helpDescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("248"))

	helpSepStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	blankPrefix    = "  "
	selectorPrefix = lipgloss.NewStyle().
			Foreground(reorderMoveColor).
			Bold(true)

	confirmBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(reorderMoveColor).
			Foreground(reorderMoveColor).
			Bold(true).
			Padding(1, 3)

	activeTabStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("110")).
			Bold(true)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(commentColor)

	appTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("110")).
			PaddingLeft(1)

	statusBarStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			PaddingLeft(1).
			PaddingRight(1)
)
