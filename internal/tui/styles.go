// styles.go defines the shared color palette and Lip Gloss styles used
// across all TUI screens. All visual constants live here so that the
// look-and-feel can be updated in one place.
package tui

import "github.com/charmbracelet/lipgloss"

// Color palette used throughout the TUI.
const (
	// ColorPurple is the primary accent color, used for titles and highlights.
	ColorPurple = lipgloss.Color("#7D56F4")
	// ColorGray is for muted, secondary text such as help hints.
	ColorGray = lipgloss.Color("#626262")
	// ColorGreen indicates success or passing checks.
	ColorGreen = lipgloss.Color("#04B575")
	// ColorRed indicates errors or failing checks.
	ColorRed = lipgloss.Color("#FF4672")
	// ColorYellow indicates warnings or caution.
	ColorYellow = lipgloss.Color("#FFCC00")
	// ColorWhite is used for primary body text.
	ColorWhite = lipgloss.Color("#FAFAFA")
	// ColorDimWhite is used for less prominent body text.
	ColorDimWhite = lipgloss.Color("#AAAAAA")
)

// BannerStyle renders the application title in bold purple.
var BannerStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(ColorPurple)

// HelpStyle renders help text in muted gray.
var HelpStyle = lipgloss.NewStyle().
	Foreground(ColorGray)

// TitleStyle renders screen titles with bold text and a bottom border.
var TitleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(ColorPurple).
	MarginBottom(1)

// SuccessStyle renders success messages in green.
var SuccessStyle = lipgloss.NewStyle().
	Foreground(ColorGreen)

// ErrorStyle renders error messages in red.
var ErrorStyle = lipgloss.NewStyle().
	Foreground(ColorRed)

// WarningStyle renders warning messages in yellow.
var WarningStyle = lipgloss.NewStyle().
	Foreground(ColorYellow)

// BodyStyle renders regular body text.
var BodyStyle = lipgloss.NewStyle().
	Foreground(ColorWhite)

// DimStyle renders de-emphasized text.
var DimStyle = lipgloss.NewStyle().
	Foreground(ColorDimWhite)

// AppStyle wraps the entire application view with consistent padding.
var AppStyle = lipgloss.NewStyle().
	Padding(1, 2)
