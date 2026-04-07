package style

import "charm.land/lipgloss/v2"

var (
	FormLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#d4d4d4"))
	FormLabelSubStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#737373"))
	FormErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f87171"))


	InputFocusedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#171717")).
		Background(lipgloss.Color("#f5f5f5"))
	InputBlurredStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#171717")).
		Background(lipgloss.Color("#a3a3a3"))
)
