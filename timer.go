package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type sessionType int

const (
	work sessionType = iota
	shortBreak
	longBreak
)

type TimerModel struct {
	timer               timer.Model
	sessionType         sessionType
	sessionCount        int
	keys                TimerKeyMap
	isRunning           bool
	customDuration      *time.Duration
	customShortBreak    *time.Duration
	customLongBreak     *time.Duration
	progressLines       int
}

type TimerKeyMap struct {
	Start key.Binding
	Reset key.Binding
	End   key.Binding
}

func DefaultTimerKeys() TimerKeyMap {
	return TimerKeyMap{
		Start: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "start/pause"),
		),
		Reset: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "reset"),
		),
		End: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "end session"),
		),
	}
}

func NewTimerModel() TimerModel {
	return NewTimerModelWithDuration(25 * time.Minute)
}

func NewTimerModelWithDuration(duration time.Duration) TimerModel {
	return NewTimerModelWithDurations(duration, 5*time.Minute, 15*time.Minute)
}

func NewTimerModelWithDurations(sessionDuration, shortBreakDuration, longBreakDuration time.Duration) TimerModel {
	return NewTimerModelWithOptions(sessionDuration, shortBreakDuration, longBreakDuration, 5)
}

func NewTimerModelWithOptions(sessionDuration, shortBreakDuration, longBreakDuration time.Duration, lines int) TimerModel {
	return TimerModel{
		timer:            timer.NewWithInterval(sessionDuration, time.Second),
		sessionType:      work,
		sessionCount:     0,
		keys:             DefaultTimerKeys(),
		isRunning:        false,
		customDuration:   &sessionDuration,
		customShortBreak: &shortBreakDuration,
		customLongBreak:  &longBreakDuration,
		progressLines:    lines,
	}
}

func (m TimerModel) Init() tea.Cmd {
	return nil
}

func (m TimerModel) Update(msg tea.Msg) (TimerModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case " ":
			if m.isRunning {
				m.isRunning = false
				return m, m.timer.Stop()
			} else {
				m.isRunning = true
				return m, m.timer.Start()
			}
		case "r":
			m.isRunning = false
			duration := m.getCurrentSessionDuration()
			m.timer = timer.NewWithInterval(duration, time.Second)
			return m, nil
		case "e":
			m.isRunning = false
			newModel := m.nextSession()
			return newModel, nil
		}
	case timer.TickMsg:
		if m.isRunning {
			var cmd tea.Cmd
			m.timer, cmd = m.timer.Update(msg)
			return m, cmd
		}
		return m, nil
	case timer.TimeoutMsg:
		m.isRunning = true
		newModel := m.nextSession()
		return newModel, newModel.timer.Start()
	}

	var cmd tea.Cmd
	m.timer, cmd = m.timer.Update(msg)
	return m, cmd
}

func (m TimerModel) IsRunning() bool {
	return m.isRunning && !m.timer.Timedout()
}

func (m *TimerModel) getSandTimer(width int) string {
	if width < 10 {
		return ""
	}
	
	remaining := m.timer.Timeout
	total := m.getCurrentSessionDuration()
	elapsed := total - remaining
	
	lines := []string{}
	
	// Calculate total elapsed seconds for global positioning
	totalSecondsElapsed := int(elapsed.Seconds())
	totalSessionSeconds := int(total.Seconds())
	
	// Create lines based on progressLines setting
	for lineIdx := 0; lineIdx < m.progressLines; lineIdx++ {
		
		var lineContent strings.Builder
		
		// Build each character in the line
		for charIdx := 0; charIdx < width; charIdx++ {
			// Calculate the global position (0 = top-right, max = bottom-left)
			// For top-right start: reverse charIdx within each line
			reversedCharIdx := width - 1 - charIdx
			globalCharPos := (lineIdx * width) + reversedCharIdx
			totalChars := m.progressLines * width
			
			// Check if this character should be dimmed based on elapsed time
			// We drain from top-right to bottom-left
			charThreshold := float64(globalCharPos) / float64(totalChars-1)
			timeThreshold := charThreshold * float64(totalSessionSeconds)
			
			// Calculate gradient color based on original position
			progress := float64(globalCharPos) / float64(totalChars-1)
			color := m.interpolateColor("#FF7CCB", "#FDFF8C", progress)
			
			if float64(totalSecondsElapsed) > timeThreshold {
				// This character is dimmed (elapsed) - use dimmed version of the gradient color
				styledChar := lipgloss.NewStyle().
					Foreground(lipgloss.Color(color)).
					Render("░")
				lineContent.WriteString(styledChar)
			} else {
				// This character is still full (remaining)
				styledChar := lipgloss.NewStyle().
					Foreground(lipgloss.Color(color)).
					Render("▓")
				lineContent.WriteString(styledChar)
			}
		}
		
		lines = append(lines, lineContent.String())
	}
	
	return strings.Join(lines, "\n")
}

func (m *TimerModel) interpolateColor(startColor, endColor string, progress float64) string {
	// Parse hex colors
	start := m.parseHexColor(startColor)
	end := m.parseHexColor(endColor)
	
	// Interpolate RGB values
	r := int(float64(start[0]) + progress*(float64(end[0])-float64(start[0])))
	g := int(float64(start[1]) + progress*(float64(end[1])-float64(start[1])))
	b := int(float64(start[2]) + progress*(float64(end[2])-float64(start[2])))
	
	// Clamp values
	if r < 0 { r = 0 }
	if r > 255 { r = 255 }
	if g < 0 { g = 0 }
	if g > 255 { g = 255 }
	if b < 0 { b = 0 }
	if b > 255 { b = 255 }
	
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

func (m *TimerModel) parseHexColor(hex string) [3]int {
	// Remove # if present
	if hex[0] == '#' {
		hex = hex[1:]
	}
	
	var r, g, b int
	fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	return [3]int{r, g, b}
}

func (m TimerModel) getCurrentSessionDuration() time.Duration {
	switch m.sessionType {
	case work:
		if m.customDuration != nil {
			return *m.customDuration
		}
		return 25 * time.Minute
	case shortBreak:
		if m.customShortBreak != nil {
			return *m.customShortBreak
		}
		return 5 * time.Minute
	case longBreak:
		if m.customLongBreak != nil {
			return *m.customLongBreak
		}
		return 15 * time.Minute
	default:
		return 25 * time.Minute
	}
}

func (m TimerModel) nextSession() TimerModel {
	switch m.sessionType {
	case work:
		m.sessionCount++
		if m.sessionCount%4 == 0 {
			m.sessionType = longBreak
		} else {
			m.sessionType = shortBreak
		}
	case shortBreak, longBreak:
		m.sessionType = work
	}

	duration := m.getCurrentSessionDuration()
	m.timer = timer.NewWithInterval(duration, time.Second)
	return m
}

func (m TimerModel) getSessionName() string {
	switch m.sessionType {
	case work:
		return "Work"
	case shortBreak:
		return "Short Break"
	case longBreak:
		return "Long Break"
	default:
		return "Work"
	}
}

func (m TimerModel) getSessionEmoji() string {
	switch m.sessionType {
	case work:
		return "🍅"
	case shortBreak:
		return "☕"
	case longBreak:
		return "🛋️"
	default:
		return "🍅"
	}
}

func (m *TimerModel) ViewWithTodos(todos []TodoItem) string {
	width := 60 // Fixed width for consistent centering

	timerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Align(lipgloss.Center).
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2).
		MarginBottom(1).
		Width(width)

	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginBottom(1).
		Align(lipgloss.Center).
		Width(width)

	todoSummaryStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("99")).
		Border(lipgloss.NormalBorder()).
		Padding(1).
		MarginTop(1).
		Align(lipgloss.Center).
		Width(width)

	// Create timer display with sand timer
	timerText := m.timer.View()
	sandTimer := m.getSandTimer(width - 4) // Account for border padding
	timerWithSand := fmt.Sprintf("%s\n%s", timerText, sandTimer)
	timerDisplay := timerStyle.Render(timerWithSand)

	var status string
	if m.IsRunning() {
		status = "Running ⏱️"
	} else {
		status = "Paused ⏸️"
	}

	statusInfo := statusStyle.Render(fmt.Sprintf("%s | Sessions: %d", status, m.sessionCount))

	// Create todo summary
	var todoSummary string
	if len(todos) == 0 {
		todoSummary = "No todos yet (Press Tab to add some)"
	} else {
		var recentTodos []string
		for i, todo := range todos {
			if i < 3 { // Show first 3 todos
				recentTodos = append(recentTodos, fmt.Sprintf("• %s", todo.Text))
			}
		}

		var todoContent string
		if len(recentTodos) > 0 {
			todoContent = lipgloss.JoinVertical(lipgloss.Left, recentTodos...)
			if len(todos) > 3 {
				todoContent += fmt.Sprintf("\n... and %d more", len(todos)-3)
			}
		}
		todoSummary = todoContent
	}

	todoSummaryDisplay := todoSummaryStyle.Render(todoSummary)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		timerDisplay,
		statusInfo,
		todoSummaryDisplay,
	)
}

func (m TimerModel) View() string {
	return m.ViewWithTodos([]TodoItem{})
}
