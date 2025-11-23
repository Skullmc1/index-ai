package main

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func runNormalRoute(path string) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(1 * time.Second)
		return doneMsg("Normal route finished (Placeholder)")
	}
}
