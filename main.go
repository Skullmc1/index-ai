package main

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	borderColor = lipgloss.Color("63")
	accentColor = lipgloss.Color("205")
	textColor   = lipgloss.Color("252")

	appStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(1, 2).
			Margin(1, 1)

	titleStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true).
			Padding(0, 1).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(borderColor)

	keyStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true)

	textStyle = lipgloss.NewStyle().
			Foreground(textColor)
)

type sessionState int

const (
	statePreloading sessionState = iota
	stateQuestion
	stateAPIKeyInput
	statePathInput
	stateRunning
	stateDone
)

type mainModel struct {
	state       sessionState
	spinner     spinner.Model
	textInput   textinput.Model
	useAI       bool
	apiKey      string
	targetPath  string
	loadingTime int
	logs        []string
	finalMsg    string
	err         error
}

type tickMsg time.Time
type progressMsg string
type doneMsg string

func initialModel() mainModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(accentColor)

	ti := textinput.New()
	ti.Placeholder = "Enter path..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 40

	return mainModel{
		state:       statePreloading,
		spinner:     s,
		textInput:   ti,
		loadingTime: 0,
		logs:        []string{},
	}
}

func (m mainModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, tickCmd())
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.state == stateDone {
			return m, tea.Quit
		}

		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "enter":
			if m.state == stateQuestion {
				m.useAI = true
				m.state = stateAPIKeyInput
				m.textInput.Placeholder = "Enter Gemini API Key"
				m.textInput.Reset()
				m.textInput.EchoMode = textinput.EchoPassword
				return m, nil
			} else if m.state == stateAPIKeyInput {
				m.apiKey = m.textInput.Value()
				m.state = statePathInput
				m.textInput.Placeholder = "Enter folder path (default: ./)"
				m.textInput.Reset()
				m.textInput.EchoMode = textinput.EchoNormal
				return m, nil
			} else if m.state == statePathInput {
				m.targetPath = m.textInput.Value()
				if m.targetPath == "" {
					dir, _ := os.Getwd()
					m.targetPath = dir
				}
				m.state = stateRunning
				if m.useAI {
					return m, runAIRoute(m.targetPath, m.apiKey)
				}
				return m, runNormalRoute(m.targetPath)
			}
		case "n", "N":
			if m.state == stateQuestion {
				m.useAI = false
				m.state = statePathInput
				m.textInput.Placeholder = "Enter folder path (default: ./)"
				m.textInput.Reset()
				m.textInput.EchoMode = textinput.EchoNormal
				return m, nil
			}
		}

	case tickMsg:
		if m.state == statePreloading {
			m.loadingTime++
			if m.loadingTime >= 2 {
				m.state = stateQuestion
				return m, nil
			}
			return m, tickCmd()
		}

	case progressMsg:
		m.logs = append(m.logs, string(msg))
		if len(m.logs) > 10 {
			m.logs = m.logs[1:]
		}
		return m, nil

	case doneMsg:
		m.state = stateDone
		m.finalMsg = string(msg)
		return m, nil

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	if m.state == stateAPIKeyInput || m.state == statePathInput {
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m mainModel) View() string {
	var content string

	switch m.state {
	case statePreloading:
		content = fmt.Sprintf("\n %s %s\n", m.spinner.View(), textStyle.Render("Initializing Core..."))

	case stateQuestion:
		content = fmt.Sprintf("\n%s\n\n%s  Yes (Gemini AI)\n%s  No (Local Logic)\n",
			textStyle.Render("Select Organization Method:"),
			keyStyle.Render("[ Enter ]"),
			keyStyle.Render("[ N ]    "))

	case stateAPIKeyInput:
		content = fmt.Sprintf("\n%s\n\n%s\n",
			textStyle.Render("Enter your Gemini API Key:"),
			m.textInput.View())

	case statePathInput:
		content = fmt.Sprintf("\n%s\n\n%s\n",
			textStyle.Render("Target Directory to Organize:"),
			m.textInput.View())

	case stateRunning:
		logView := ""
		for _, log := range m.logs {
			logView += fmt.Sprintf("%s %s\n", keyStyle.Render(">"), textStyle.Render(log))
		}
		content = fmt.Sprintf("\n %s %s\n\n%s", m.spinner.View(), textStyle.Render("Processing..."), logView)

	case stateDone:
		color := accentColor
		if len(m.finalMsg) > 5 && (m.finalMsg[:5] == "Error" || m.finalMsg[:6] == "Failed") {
			color = lipgloss.Color("196")
		}

		resultStyle := lipgloss.NewStyle().Foreground(color).Bold(true).Width(50)
		content = fmt.Sprintf("\n%s\n\n%s\n", resultStyle.Render(m.finalMsg), textStyle.Render("[ Press any key to exit ]"))
	}

	title := titleStyle.Render("INDEX AI")
	ui := lipgloss.JoinVertical(lipgloss.Left, title, content)

	return appStyle.Render(ui)
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
