package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type viewState int

const (
	timerView viewState = iota
	todoView
)

type model struct {
	timer  TimerModel
	todo   TodoModel
	view   viewState
	keys   KeyMap
	width  int
	height int
}

func initialModel(sessionDuration, shortBreakDuration, longBreakDuration time.Duration, lines int) model {
	return model{
		timer: NewTimerModelWithOptions(sessionDuration, shortBreakDuration, longBreakDuration, lines),
		todo:  NewTodoModel(),
		view:  timerView,
		keys:  DefaultKeyMap(),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.timer.Init(), m.todo.Init())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab":
			if m.view == timerView {
				m.view = todoView
			} else {
				m.view = timerView
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	var timerCmd tea.Cmd

	// Always update timer to handle tick messages
	m.timer, timerCmd = m.timer.Update(msg)

	if m.view == timerView {
		cmd = timerCmd
	} else {
		m.todo, cmd = m.todo.Update(msg)
		// Still need to handle timer commands even in todo view
		if timerCmd != nil {
			cmd = tea.Batch(cmd, timerCmd)
		}
	}

	return m, cmd
}

func (m model) View() string {
	width := 60 // Fixed width for consistent centering

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Align(lipgloss.Center).
		Width(width)

	var content string
	if m.view == timerView {
		content = m.timer.ViewWithTodos(m.todo.todos)
	} else {
		content = m.todo.View()
	}

	help := helpStyle.Render(m.keys.ShortHelp())

	// Create the main content area
	mainContent := lipgloss.JoinVertical(
		lipgloss.Center,
		content,
		"",
		help,
	)

	// Center the entire content
	containerStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center)

	return containerStyle.Render(mainContent)
}

func main() {
	sessionFlag := flag.String("s", "25m", "Session duration (e.g., 10m, 1h30m)")
	shortBreakFlag := flag.String("sb", "5m", "Short break duration (e.g., 5m, 10m)")
	longBreakFlag := flag.String("lb", "15m", "Long break duration (e.g., 15m, 30m)")
	linesFlag := flag.Int("l", 5, "Number of progress bar lines")
	flag.Parse()
	
	sessionDuration, err := time.ParseDuration(*sessionFlag)
	if err != nil {
		fmt.Printf("Error parsing session duration: %v\n", err)
		fmt.Println("Examples: 10m, 25m, 1h, 1h30m")
		os.Exit(1)
	}
	
	shortBreakDuration, err := time.ParseDuration(*shortBreakFlag)
	if err != nil {
		fmt.Printf("Error parsing short break duration: %v\n", err)
		fmt.Println("Examples: 5m, 10m, 15m")
		os.Exit(1)
	}
	
	longBreakDuration, err := time.ParseDuration(*longBreakFlag)
	if err != nil {
		fmt.Printf("Error parsing long break duration: %v\n", err)
		fmt.Println("Examples: 15m, 30m, 45m")
		os.Exit(1)
	}
	
	p := tea.NewProgram(initialModel(sessionDuration, shortBreakDuration, longBreakDuration, *linesFlag), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
