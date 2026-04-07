// target.go implements the install target selection screen.
// Users can choose whether STUI should operate on the local host
// or connect to a remote machine over SSH before preflight begins.
package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/EtherAura/stui/internal/installer"
)

// StartPreflightMsg signals the app to transition from the target
// screen to the preflight check screen.
type StartPreflightMsg struct {
	// AppID is the registry key of the selected application.
	AppID string
	// Target is the chosen install target.
	Target installer.Target
}

type targetField struct {
	key         string
	label       string
	placeholder string
}

var targetFields = []targetField{
	{key: "mode", label: "Install Target Mode", placeholder: "local or ssh"},
	{key: "user", label: "SSH Username", placeholder: "ubuntu"},
	{key: "host", label: "SSH Host", placeholder: "192.168.0.10"},
	{key: "port", label: "SSH Port", placeholder: "22"},
}

// TargetModel collects the install target before preflight starts.
type TargetModel struct {
	appID      string
	fields     []targetField
	inputs     []textinput.Model
	focusIndex int
	err        string
	width      int
	height     int
}

// NewTargetModel creates a target selection screen for the given app.
func NewTargetModel(appID string) TargetModel {
	inputs := make([]textinput.Model, len(targetFields))
	for i, f := range targetFields {
		ti := textinput.New()
		ti.Placeholder = f.placeholder
		ti.CharLimit = 256
		ti.Width = 32
		switch f.key {
		case "mode":
			ti.SetValue(string(installer.TargetModeLocal))
		case "port":
			ti.SetValue("22")
		}
		if i == 0 {
			ti.Focus()
		}
		inputs[i] = ti
	}

	return TargetModel{
		appID:  appID,
		fields: targetFields,
		inputs: inputs,
	}
}

// Init implements tea.Model.
func (m TargetModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements tea.Model.
func (m TargetModel) Update(msg tea.Msg) (TargetModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg {
				return BackToMenuMsg{}
			}
		case "tab", "down":
			return m.nextField()
		case "shift+tab", "up":
			return m.prevField()
		case "enter":
			mode := strings.ToLower(strings.TrimSpace(m.inputs[0].Value()))
			if mode == "" || mode == string(installer.TargetModeLocal) || m.focusIndex == len(m.inputs)-1 {
				return m.submit()
			}
			return m.nextField()
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	var cmd tea.Cmd
	m.inputs[m.focusIndex], cmd = m.inputs[m.focusIndex].Update(msg)
	return m, cmd
}

func (m TargetModel) nextField() (TargetModel, tea.Cmd) {
	m.inputs[m.focusIndex].Blur()
	m.focusIndex = (m.focusIndex + 1) % len(m.inputs)
	return m, m.inputs[m.focusIndex].Focus()
}

func (m TargetModel) prevField() (TargetModel, tea.Cmd) {
	m.inputs[m.focusIndex].Blur()
	m.focusIndex = (m.focusIndex - 1 + len(m.inputs)) % len(m.inputs)
	return m, m.inputs[m.focusIndex].Focus()
}

func (m TargetModel) submit() (TargetModel, tea.Cmd) {
	target, err := m.buildTarget()
	if err != nil {
		m.err = err.Error()
		return m, nil
	}

	return m, func() tea.Msg {
		return StartPreflightMsg{AppID: m.appID, Target: target}
	}
}

func (m TargetModel) buildTarget() (installer.Target, error) {
	mode := installer.TargetMode(strings.ToLower(strings.TrimSpace(m.inputs[0].Value())))
	if mode == "" {
		mode = installer.TargetModeLocal
	}

	target := installer.Target{Mode: mode}
	if mode == installer.TargetModeSSH {
		target.User = strings.TrimSpace(m.inputs[1].Value())
		target.Host = strings.TrimSpace(m.inputs[2].Value())

		portText := strings.TrimSpace(m.inputs[3].Value())
		if portText != "" {
			port, err := strconv.Atoi(portText)
			if err != nil {
				return installer.Target{}, fmt.Errorf("SSH port must be a number")
			}
			target.Port = port
		}
	}

	if err := target.Validate(); err != nil {
		return installer.Target{}, err
	}
	return target, nil
}

// View implements tea.Model.
func (m TargetModel) View() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Install Target"))
	b.WriteString("\n")
	b.WriteString(DimStyle.Render("Use `local` for this machine or `ssh` for a remote Ubuntu host."))
	b.WriteString("\n\n")

	mode := strings.ToLower(strings.TrimSpace(m.inputs[0].Value()))
	for i, f := range m.fields {
		if mode != string(installer.TargetModeSSH) && i > 0 {
			continue
		}

		if i == m.focusIndex {
			b.WriteString(BannerStyle.Render(f.label))
		} else {
			b.WriteString(BodyStyle.Render(f.label))
		}
		b.WriteString("\n")
		b.WriteString("  " + m.inputs[i].View())
		b.WriteString("\n\n")
	}

	if m.err != "" {
		b.WriteString(ErrorStyle.Render("✗ " + m.err))
		b.WriteString("\n\n")
	}

	b.WriteString(HelpStyle.Render("Press enter to continue, tab to move, esc to go back"))
	return AppStyle.Render(b.String())
}

// AppID returns the selected application id.
func (m TargetModel) AppID() string {
	return m.appID
}

// FocusIndex returns the currently focused field index.
func (m TargetModel) FocusIndex() int {
	return m.focusIndex
}
