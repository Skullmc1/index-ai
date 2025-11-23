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
	finalMsg    string // Added to store the result message
	err         error
}

type tickMsg time.Time
type progressMsg string
type doneMsg string

func initialModel() mainModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	ti := textinput.New()
	ti.Placeholder = "Enter path..."
	ti.Focus()

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
		// If we are done, any key quits the app
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
		return m, nil

	case doneMsg:
		m.state = stateDone
		m.finalMsg = string(msg) // Save the message
		return m, nil            // Do NOT quit here; wait for user input

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
	var s string

	switch m.state {
	case statePreloading:
		s = fmt.Sprintf("\n %s Initializing Index AI...\n", m.spinner.View())

	case stateQuestion:
		s = "\nDo you want to use AI features?\n\n[ Enter ] Yes (Gemini)\n[ N ] No (Normal)\n"

	case stateAPIKeyInput:
		s = fmt.Sprintf("\nInput your API Key:\n\n%s\n", m.textInput.View())

	case statePathInput:
		s = fmt.Sprintf("\nWhere should we organize?\n\n%s\n", m.textInput.View())

	case stateRunning:
		s = fmt.Sprintf("\n %s Processing...\n\n", m.spinner.View())
		for _, log := range m.logs {
			s += fmt.Sprintf("> %s\n", log)
		}

	case stateDone:
		// Display the final message and wait
		color := "205" // Pink for success
		if len(m.finalMsg) > 5 && m.finalMsg[:5] == "Error" || m.finalMsg[:6] == "Failed" {
			color = "196" // Red for error
		}

		style := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(true)
		s = fmt.Sprintf("\n%s\n\n[ Press any key to exit ]\n", style.Render(m.finalMsg))
	}

	return lipgloss.NewStyle().Margin(1, 2).Render(s)
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
