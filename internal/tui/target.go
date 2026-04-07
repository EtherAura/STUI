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
	{key: "password", label: "SSH Password", placeholder: "password"},
	{key: "key_path", label: "SSH Key Path", placeholder: "~/.ssh/id_ed25519"},
}

// TargetModel collects the install target before preflight starts.
type TargetModel struct {
	appID      string
	mode       installer.TargetMode
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
		case "port":
			ti.SetValue("22")
		case "password":
			ti.EchoMode = textinput.EchoPassword
			ti.EchoCharacter = '•'
		}
		if i == 0 {
			// Focus is visual-only for the mode selector.
		}
		inputs[i] = ti
	}

	return TargetModel{
		appID:  appID,
		mode:   installer.TargetModeLocal,
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
		case "left":
			if m.focusIndex == 1 {
				m.setMode(installer.TargetModeLocal)
				return m, nil
			}
			if m.focusIndex == 2 {
				m.setMode(installer.TargetModeSSH)
				return m, nil
			}
		case "right":
			if m.focusIndex == 1 {
				m.setMode(installer.TargetModeLocal)
				return m, nil
			}
			if m.focusIndex == 2 {
				m.setMode(installer.TargetModeSSH)
				return m, nil
			}
		case " ":
			if m.isOptionRow() {
				m.selectFocusedOption()
				return m, nil
			}
		case "enter":
			if m.isOptionRow() {
				m.selectFocusedOption()
				return m, nil
			}
			if m.isProceedRow() {
				return m.submit()
			}
			return m.nextField()
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	var cmd tea.Cmd
	if !m.isInputRow() {
		return m, nil
	}
	inputIndex := m.inputIndex()
	m.inputs[inputIndex], cmd = m.inputs[inputIndex].Update(msg)
	return m, cmd
}

func (m TargetModel) nextField() (TargetModel, tea.Cmd) {
	if m.isInputRow() {
		m.inputs[m.inputIndex()].Blur()
	}
	m.focusIndex++
	if m.focusIndex >= m.rowCount() {
		m.focusIndex = 1
	}
	if !m.isInputRow() {
		return m, nil
	}
	return m, m.inputs[m.inputIndex()].Focus()
}

func (m TargetModel) prevField() (TargetModel, tea.Cmd) {
	if m.isInputRow() {
		m.inputs[m.inputIndex()].Blur()
	}
	m.focusIndex--
	if m.focusIndex < 1 {
		m.focusIndex = m.rowCount() - 1
	}
	if !m.isInputRow() {
		return m, nil
	}
	return m, m.inputs[m.inputIndex()].Focus()
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
	target := installer.Target{Mode: m.mode}
	if m.mode == installer.TargetModeSSH {
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
		target.Password = m.inputs[4].Value()
		target.KeyPath = strings.TrimSpace(m.inputs[5].Value())
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
	b.WriteString(DimStyle.Render("Choose where STUI should run preflight and installation steps."))
	b.WriteString("\n\n")

	b.WriteString(BodyStyle.Render("Install Target Mode"))
	b.WriteString("\n")
	b.WriteString("  " + m.modeOption("Local", installer.TargetModeLocal, m.focusIndex == 1))
	b.WriteString("\n")
	b.WriteString("  " + m.modeOption("Remote", installer.TargetModeSSH, m.focusIndex == 2))
	b.WriteString("\n\n")

	if m.mode == installer.TargetModeSSH {
		b.WriteString(DimStyle.Render("Provide an SSH password, a private key path, or leave both empty to use agent/default keys."))
		b.WriteString("\n\n")
		for i, f := range m.fields[1:] {
			rowIndex := i + 3
			if rowIndex == m.focusIndex {
				b.WriteString(BannerStyle.Render(f.label))
			} else {
				b.WriteString(BodyStyle.Render(f.label))
			}
			b.WriteString("\n")
			b.WriteString("  " + m.inputs[i+1].View())
			b.WriteString("\n\n")
		}
	}

	proceedFocused := m.isProceedRow()
	if proceedFocused {
		b.WriteString(BannerStyle.Render("[ Proceed ]"))
	} else {
		b.WriteString(BodyStyle.Render("[ Proceed ]"))
	}
	b.WriteString("\n\n")

	if m.err != "" {
		b.WriteString(ErrorStyle.Render("✗ " + m.err))
		b.WriteString("\n\n")
	}

	b.WriteString(HelpStyle.Render("Use up/down to move, enter or space to choose an option, enter on Proceed to continue, esc to go back"))
	return AppStyle.Render(b.String())
}

// AppID returns the selected application id.
func (m TargetModel) AppID() string {
	return m.appID
}

// FocusIndex returns the currently focused row index.
func (m TargetModel) FocusIndex() int {
	return m.focusIndex
}

func (m *TargetModel) setMode(mode installer.TargetMode) {
	m.mode = mode
	if mode == installer.TargetModeLocal && m.focusIndex > 2 {
		m.focusIndex = m.rowCount() - 1
	}
}

func (m *TargetModel) selectFocusedOption() {
	switch m.focusIndex {
	case 1:
		m.setMode(installer.TargetModeLocal)
	case 2:
		m.setMode(installer.TargetModeSSH)
	}
}

func (m TargetModel) rowCount() int {
	if m.mode == installer.TargetModeSSH {
		return 9
	}
	return 4
}

func (m TargetModel) isOptionRow() bool {
	return m.focusIndex == 1 || m.focusIndex == 2
}

func (m TargetModel) isInputRow() bool {
	return m.mode == installer.TargetModeSSH && m.focusIndex >= 3 && m.focusIndex <= 7
}

func (m TargetModel) isProceedRow() bool {
	return m.focusIndex == m.rowCount()-1
}

func (m TargetModel) inputIndex() int {
	switch m.focusIndex {
	case 3:
		return 1
	case 4:
		return 2
	case 5:
		return 3
	case 6:
		return 4
	case 7:
		return 5
	default:
		return 0
	}
}

func (m TargetModel) modeOption(label string, mode installer.TargetMode, focused bool) string {
	text := label
	if m.mode == mode {
		text = "(*) " + text
	} else {
		text = "( ) " + text
	}

	if focused && m.mode == mode {
		return SuccessStyle.Render(text)
	}
	if focused {
		return BannerStyle.Render(text)
	}
	if m.mode == mode {
		return SuccessStyle.Render(text)
	}
	return DimStyle.Render(text)
}
